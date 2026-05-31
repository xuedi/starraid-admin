// Command seed authors the minimal starting world into PostgreSQL (see
// docs/admin.md, docs/database.md). It is the "admin → DB (authored content)"
// path: it writes instances composed from the catalog the server already
// migrated — it never invents an object type. Idempotent: re-running yields the
// same starting area (1 sector, 1 account, 1 character, 4 ships) with no dupes.
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
	playerShip    = "Test Pilot's Ship"
)

// npcShips are the three unowned ships and their spread-out positions.
var npcShips = []struct {
	name string
	x, y int64
}{
	{"Marauder", 5000, 0},
	{"Drifter", 0, 5000},
	{"Sentinel", -5000, -5000},
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

	// Object type — resolve from the catalog the server migrated; copy its base
	// stats into each instance (admin never invents a type).
	var typeID int
	var baseHealth, baseShield, baseScanner, baseJammer int
	if err = tx.QueryRow(ctx,
		`SELECT id, base_health, base_shield, base_scanner, base_jammer
		   FROM object_types WHERE key = 'spaceship'`,
	).Scan(&typeID, &baseHealth, &baseShield, &baseScanner, &baseJammer); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("object_types has no 'spaceship' — run the server migrations first")
		}
		return fmt.Errorf("lookup spaceship type: %w", err)
	}

	// Objects — delete this sector's existing objects, then insert the 4 fresh so
	// re-seeding is deterministic (count stays 4, no duplicates).
	if _, err = tx.Exec(ctx, `DELETE FROM objects WHERE sector_id = $1`, sectorID); err != nil {
		return fmt.Errorf("clear sector objects: %w", err)
	}

	insertObject := func(name string, ownerCharacterID *int64, x, y int64) (int64, error) {
		var id int64
		err := tx.QueryRow(ctx,
			`INSERT INTO objects
			   (type_id, owner_character_id, sector_id, name, x, y, health, shield, scanner, jammer)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 RETURNING id`,
			typeID, ownerCharacterID, sectorID, name, x, y,
			baseHealth, baseShield, baseScanner, baseJammer,
		).Scan(&id)
		return id, err
	}

	objectIDs := make([]int64, 0, 4)
	playerID, err := insertObject(playerShip, &characterID, 0, 0)
	if err != nil {
		return fmt.Errorf("insert player ship: %w", err)
	}
	objectIDs = append(objectIDs, playerID)
	for _, s := range npcShips {
		id, err := insertObject(s.name, nil, s.x, s.y)
		if err != nil {
			return fmt.Errorf("insert npc ship %q: %w", s.name, err)
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
		"object_ids", objectIDs,
	)
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
