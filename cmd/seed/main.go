// Command seed authors the starting world into PostgreSQL (see docs/admin.md,
// docs/database.md). It is the "admin → DB (authored content)" path: it reads the
// catalog the server synced (object_class / module_types / item_types) and writes
// fat INSTANCES composed from it — picking a class, then referencing modules into
// slots and items into cargo. It never invents a class/module/item; the catalog
// is the contract. The class says what FITS, the seed authors what is FITTED.
//
// Idempotent: re-running deletes this sector's objects (cascading to their
// modules/cargo) and re-inserts the same scene — a varied starting area (player
// skiff + mixed NPC ships + a station + asteroids).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	sectorName    = "Starting Area"
	testEmail     = "test@example.org"
	testPassword  = "1234"
	characterName = "Test Pilot"
)

// shipModules is the standard ship fitting the seed authors: generator + shield +
// thruster (all internal). Stations drop the thruster; asteroids carry none.
var shipModules = []string{"gen_hydrogen_basic", "shield_basic", "thruster_ion"}

// cargoSpec is one cargo stack to place in an instance's base hold.
type cargoSpec struct {
	item string
	qty  int64
}

// placement is one object in the starting scene: a class, a name, an optional
// owner (nil = NPC/structure/asteroid), a position, and its authored fitting.
type placement struct {
	class   string
	name    string
	owned   bool // true → owned by the test character (the player ship)
	x, y    int64
	modules []string
	cargo   []cargoSpec
}

// scene is the varied starting area: the player's skiff at the origin, mixed NPC
// ships, a station, and several asteroids spread around it.
var scene = []placement{
	{class: "skiff", name: "Test Pilot's Ship", owned: true, x: 0, y: 0,
		modules: shipModules, cargo: []cargoSpec{{"hydrogen", 100}}},
	{class: "corvette", name: "Marauder", x: 6000, y: 0,
		modules: shipModules, cargo: []cargoSpec{{"hydrogen", 200}, {"ammunition", 50}}},
	{class: "hauler", name: "Drifter", x: 0, y: 6000,
		modules: shipModules, cargo: []cargoSpec{{"hydrogen", 500}, {"iron_ore", 1000}}},
	{class: "cruiser", name: "Sentinel", x: -6000, y: -6000,
		modules: shipModules, cargo: []cargoSpec{{"hydrogen", 300}}},
	{class: "station", name: "Waystation", x: 8000, y: 8000,
		modules: []string{"gen_hydrogen_basic", "shield_basic"}, cargo: []cargoSpec{{"hydrogen", 2000}}},
	{class: "asteroid", name: "Asteroid Alpha", x: 3000, y: -3000,
		cargo: []cargoSpec{{"iron_ore", 5000}, {"ice", 2000}}},
	{class: "asteroid", name: "Asteroid Beta", x: -4000, y: 2000,
		cargo: []cargoSpec{{"iron_ore", 3000}}},
	{class: "asteroid", name: "Asteroid Gamma", x: 9000, y: -2000,
		cargo: []cargoSpec{{"ice", 4000}}},
}

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	dsn := envOr("DATABASE_URL", "postgres://starraid:starraid@localhost:5432/starraid?sslmode=disable")

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful Commit

	// Sector — idempotent: reuse the existing "Starting Area" if present.
	var sectorID int64
	err = tx.QueryRow(ctx, `SELECT id FROM sectors WHERE name = $1`, sectorName).Scan(&sectorID)
	if errors.Is(err, pgx.ErrNoRows) {
		if err = tx.QueryRow(ctx,
			`INSERT INTO sectors (name) VALUES ($1) RETURNING id`, sectorName,
		).Scan(&sectorID); err != nil {
			return fmt.Errorf("insert sector: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("lookup sector: %w", err)
	}

	// Account — upsert by the unique email so the hash/status stay re-runnable.
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	var accountID int64
	if err = tx.QueryRow(ctx,
		`INSERT INTO accounts (email, password_hash, status) VALUES ($1, $2, 'active')
		   ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash,
		                                     status = EXCLUDED.status
		 RETURNING id`, testEmail, string(hash),
	).Scan(&accountID); err != nil {
		return fmt.Errorf("upsert account: %w", err)
	}

	// Character — ensure exactly one for the account (no unique constraint, so
	// guard on name).
	var characterID int64
	err = tx.QueryRow(ctx,
		`SELECT id FROM characters WHERE account_id = $1 AND name = $2`, accountID, characterName,
	).Scan(&characterID)
	if errors.Is(err, pgx.ErrNoRows) {
		if err = tx.QueryRow(ctx,
			`INSERT INTO characters (account_id, name) VALUES ($1, $2) RETURNING id`,
			accountID, characterName,
		).Scan(&characterID); err != nil {
			return fmt.Errorf("insert character: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("lookup character: %w", err)
	}

	// Read the catalog the server synced — the admin's contract. It learns which
	// classes/modules/items exist from the DB, never from the server's Go code.
	cat, err := loadCatalog(ctx, tx)
	if err != nil {
		return err
	}

	// Objects — delete this sector's existing objects (cascades to object_modules
	// and object_items), then author the scene fresh so re-seeding is deterministic.
	if _, err = tx.Exec(ctx, `DELETE FROM objects WHERE sector_id = $1`, sectorID); err != nil {
		return fmt.Errorf("clear sector objects: %w", err)
	}

	objectIDs := make([]int64, 0, len(scene))
	for _, p := range scene {
		var owner *int64
		if p.owned {
			owner = &characterID
		}
		id, err := cat.fit(ctx, tx, sectorID, owner, p)
		if err != nil {
			return fmt.Errorf("fit %q (%s): %w", p.name, p.class, err)
		}
		objectIDs = append(objectIDs, id)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	slog.Info("seeded starting world",
		"sector_id", sectorID,
		"account_id", accountID,
		"character_id", characterID,
		"objects", len(objectIDs),
		"object_ids", objectIDs,
	)
	return nil
}

// catalog is the lookup the seed composes instances from: class id by key, module
// (id + slot kind) by key, item id by key — all read from the synced catalog.
type catalog struct {
	classID map[string]int
	module  map[string]moduleRef
	itemID  map[string]int
}

type moduleRef struct {
	id       int
	slotKind string
}

func loadCatalog(ctx context.Context, tx pgx.Tx) (*catalog, error) {
	c := &catalog{classID: map[string]int{}, module: map[string]moduleRef{}, itemID: map[string]int{}}

	rows, err := tx.Query(ctx, `SELECT key, id FROM object_class`)
	if err != nil {
		return nil, fmt.Errorf("load classes: %w", err)
	}
	for rows.Next() {
		var key string
		var id int
		if err := rows.Scan(&key, &id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan class: %w", err)
		}
		c.classID[key] = id
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = tx.Query(ctx, `SELECT key, id, slot_kind FROM module_types`)
	if err != nil {
		return nil, fmt.Errorf("load modules: %w", err)
	}
	for rows.Next() {
		var key, slotKind string
		var id int
		if err := rows.Scan(&key, &id, &slotKind); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan module: %w", err)
		}
		c.module[key] = moduleRef{id: id, slotKind: slotKind}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = tx.Query(ctx, `SELECT key, id FROM item_types`)
	if err != nil {
		return nil, fmt.Errorf("load items: %w", err)
	}
	for rows.Next() {
		var key string
		var id int
		if err := rows.Scan(&key, &id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan item: %w", err)
		}
		c.itemID[key] = id
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(c.classID) == 0 || len(c.module) == 0 || len(c.itemID) == 0 {
		return nil, errors.New("catalog is empty — run the server migrations + catalog sync first (`just migrate`)")
	}
	return c, nil
}

// fit inserts one object instance and authors its fitting: each module referenced
// into the next free slot of its kind (honouring the class's slot layout — the
// standard fittings stay within capacity), each cargo stack into the base hold
// (module_id NULL). health/shield are left NULL: the server stamps them from the
// derived attributes on load.
func (c *catalog) fit(ctx context.Context, tx pgx.Tx, sectorID int64, owner *int64, p placement) (int64, error) {
	classID, ok := c.classID[p.class]
	if !ok {
		return 0, fmt.Errorf("unknown class %q", p.class)
	}

	var objID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO objects (object_class_id, owner_character_id, sector_id, name, x, y)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		classID, owner, sectorID, p.name, p.x, p.y,
	).Scan(&objID); err != nil {
		return 0, fmt.Errorf("insert object: %w", err)
	}

	slotIndex := map[string]int{} // next free slot index per slot kind
	for _, mk := range p.modules {
		m, ok := c.module[mk]
		if !ok {
			return 0, fmt.Errorf("unknown module %q", mk)
		}
		idx := slotIndex[m.slotKind]
		slotIndex[m.slotKind]++
		if _, err := tx.Exec(ctx,
			`INSERT INTO object_modules (object_id, module_type_id, slot_kind, slot_index)
			 VALUES ($1, $2, $3, $4)`,
			objID, m.id, m.slotKind, idx,
		); err != nil {
			return 0, fmt.Errorf("insert module %q: %w", mk, err)
		}
	}

	for _, cs := range p.cargo {
		itemID, ok := c.itemID[cs.item]
		if !ok {
			return 0, fmt.Errorf("unknown item %q", cs.item)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO object_items (object_id, module_id, item_type_id, quantity)
			 VALUES ($1, NULL, $2, $3)`,
			objID, itemID, cs.qty,
		); err != nil {
			return 0, fmt.Errorf("insert cargo %q: %w", cs.item, err)
		}
	}
	return objID, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
