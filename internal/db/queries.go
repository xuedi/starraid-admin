package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Summary returns the Dashboard counts + an objects-per-class breakdown.
func (d *DB) Summary(ctx context.Context) (Summary, error) {
	var s Summary
	err := d.pool.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM accounts),
		       (SELECT count(*) FROM characters),
		       (SELECT count(*) FROM objects),
		       (SELECT count(*) FROM sectors)`).
		Scan(&s.Accounts, &s.Characters, &s.Objects, &s.Sectors)
	if err != nil {
		return Summary{}, fmt.Errorf("summary counts: %w", err)
	}

	rows, err := d.pool.Query(ctx, `
		SELECT oc.key, oc.name, count(o.id)
		  FROM object_class oc
		  JOIN objects o ON o.object_class_id = oc.id
		 GROUP BY oc.key, oc.name
		 ORDER BY count(o.id) DESC, oc.name`)
	if err != nil {
		return Summary{}, fmt.Errorf("summary by-class: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c ClassCount
		if err := rows.Scan(&c.Key, &c.Name, &c.Count); err != nil {
			return Summary{}, fmt.Errorf("scan by-class: %w", err)
		}
		s.ByClass = append(s.ByClass, c)
	}
	if err := rows.Err(); err != nil {
		return Summary{}, err
	}
	return s, nil
}

// Users returns every account with its character count (base table details).
func (d *DB) Users(ctx context.Context) ([]UserRow, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT a.id, a.email, a.status, a.created_at, a.last_login_at,
		       count(c.id)
		  FROM accounts a
		  LEFT JOIN characters c ON c.account_id = a.id
		 GROUP BY a.id
		 ORDER BY a.id`)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}
	defer rows.Close()
	out := []UserRow{}
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.Email, &u.Status, &u.CreatedAt, &u.LastLoginAt, &u.Characters); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// User returns one account with its characters and each character's owned
// objects. Returns ErrNotFound if no account has the id.
func (d *DB) User(ctx context.Context, id int64) (*UserDetail, error) {
	var u UserDetail
	err := d.pool.QueryRow(ctx, `
		SELECT id, email, status, created_at, last_login_at
		  FROM accounts WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.Status, &u.CreatedAt, &u.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	rows, err := d.pool.Query(ctx, `
		SELECT id, name, prestige, faction, created_at
		  FROM characters WHERE account_id = $1 ORDER BY id`, id)
	if err != nil {
		return nil, fmt.Errorf("user characters: %w", err)
	}
	defer rows.Close()
	byID := map[int64]int{} // character id → index in u.Characters
	for rows.Next() {
		var c Character
		if err := rows.Scan(&c.ID, &c.Name, &c.Prestige, &c.Faction, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan character: %w", err)
		}
		c.Objects = []OwnedObject{}
		byID[c.ID] = len(u.Characters)
		u.Characters = append(u.Characters, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Owned objects for the whole account, grouped back onto their character.
	orows, err := d.pool.Query(ctx, `
		SELECT o.owner_character_id, o.id, o.name, oc.name, s.name
		  FROM objects o
		  JOIN object_class oc ON oc.id = o.object_class_id
		  JOIN sectors s ON s.id = o.sector_id
		  JOIN characters c ON c.id = o.owner_character_id
		 WHERE c.account_id = $1
		 ORDER BY o.id`, id)
	if err != nil {
		return nil, fmt.Errorf("user objects: %w", err)
	}
	defer orows.Close()
	for orows.Next() {
		var charID int64
		var o OwnedObject
		if err := orows.Scan(&charID, &o.ID, &o.Name, &o.Class, &o.Sector); err != nil {
			return nil, fmt.Errorf("scan owned object: %w", err)
		}
		if idx, ok := byID[charID]; ok {
			u.Characters[idx].Objects = append(u.Characters[idx].Objects, o)
		}
	}
	return &u, orows.Err()
}

// Objects returns the object list, optionally filtered to one sector.
func (d *DB) Objects(ctx context.Context, sectorID *int64) ([]ObjectRow, error) {
	const base = `
		SELECT o.id, o.name, oc.key, oc.name, oc.kind, s.id, s.name,
		       ch.name, o.x, o.y, o.health, o.shield, o.status
		  FROM objects o
		  JOIN object_class oc ON oc.id = o.object_class_id
		  JOIN sectors s ON s.id = o.sector_id
		  LEFT JOIN characters ch ON ch.id = o.owner_character_id`
	var (
		rows pgx.Rows
		err  error
	)
	if sectorID != nil {
		rows, err = d.pool.Query(ctx, base+` WHERE o.sector_id = $1 ORDER BY o.id`, *sectorID)
	} else {
		rows, err = d.pool.Query(ctx, base+` ORDER BY o.id`)
	}
	if err != nil {
		return nil, fmt.Errorf("objects: %w", err)
	}
	defer rows.Close()
	out := []ObjectRow{}
	for rows.Next() {
		var o ObjectRow
		if err := rows.Scan(&o.ID, &o.Name, &o.ClassKey, &o.ClassName, &o.Kind,
			&o.SectorID, &o.Sector, &o.Owner, &o.X, &o.Y, &o.Health, &o.Shield, &o.Status); err != nil {
			return nil, fmt.Errorf("scan object: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// Object returns the full detail for one object (class caps, owner, fitting,
// cargo). Returns ErrNotFound if no object has the id.
func (d *DB) Object(ctx context.Context, id int64) (*ObjectDetail, error) {
	var o ObjectDetail
	var slots []byte
	err := d.pool.QueryRow(ctx, `
		SELECT o.id, o.name, o.status, o.x, o.y, o.health, o.shield,
		       o.created_at, o.updated_at,
		       oc.key, oc.name, oc.kind, oc.size_class, oc.base_mass,
		       oc.base_cargo_volume, oc.slots,
		       s.id, s.name,
		       o.owner_character_id, ch.name, ch.account_id
		  FROM objects o
		  JOIN object_class oc ON oc.id = o.object_class_id
		  JOIN sectors s ON s.id = o.sector_id
		  LEFT JOIN characters ch ON ch.id = o.owner_character_id
		 WHERE o.id = $1`, id).
		Scan(&o.ID, &o.Name, &o.Status, &o.X, &o.Y, &o.Health, &o.Shield,
			&o.CreatedAt, &o.UpdatedAt,
			&o.Class.Key, &o.Class.Name, &o.Class.Kind, &o.Class.SizeClass, &o.Class.BaseMass,
			&o.Class.BaseCargoVolume, &slots,
			&o.SectorID, &o.Sector,
			&o.OwnerCharacterID, &o.Owner, &o.OwnerAccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("object: %w", err)
	}
	o.Class.Slots = rawOrEmpty(slots)
	o.Modules = []ModuleFitting{}
	o.Cargo = []CargoStack{}

	mrows, err := d.pool.Query(ctx, `
		SELECT om.id, mt.key, mt.name, om.slot_kind, om.slot_index,
		       om.quality, om.status, mt.params
		  FROM object_modules om
		  JOIN module_types mt ON mt.id = om.module_type_id
		 WHERE om.object_id = $1
		 ORDER BY om.slot_kind, om.slot_index`, id)
	if err != nil {
		return nil, fmt.Errorf("object modules: %w", err)
	}
	defer mrows.Close()
	for mrows.Next() {
		var m ModuleFitting
		var params []byte
		if err := mrows.Scan(&m.ID, &m.Key, &m.Name, &m.SlotKind, &m.SlotIndex,
			&m.Quality, &m.Status, &params); err != nil {
			return nil, fmt.Errorf("scan module: %w", err)
		}
		m.Params = rawOrEmpty(params)
		o.Modules = append(o.Modules, m)
	}
	if err := mrows.Err(); err != nil {
		return nil, err
	}

	crows, err := d.pool.Query(ctx, `
		SELECT oi.id, it.key, it.name, it.category, oi.quantity
		  FROM object_items oi
		  JOIN item_types it ON it.id = oi.item_type_id
		 WHERE oi.object_id = $1
		 ORDER BY it.name`, id)
	if err != nil {
		return nil, fmt.Errorf("object items: %w", err)
	}
	defer crows.Close()
	for crows.Next() {
		var c CargoStack
		if err := crows.Scan(&c.ID, &c.Key, &c.Name, &c.Category, &c.Quantity); err != nil {
			return nil, fmt.Errorf("scan cargo: %w", err)
		}
		o.Cargo = append(o.Cargo, c)
	}
	return &o, crows.Err()
}

// Sectors returns every sector with its object count.
func (d *DB) Sectors(ctx context.Context) ([]Sector, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT s.id, s.name, s.created_at,
		       (SELECT count(*) FROM objects o WHERE o.sector_id = s.id)
		  FROM sectors s ORDER BY s.id`)
	if err != nil {
		return nil, fmt.Errorf("sectors: %w", err)
	}
	defer rows.Close()
	out := []Sector{}
	for rows.Next() {
		var s Sector
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.Objects); err != nil {
			return nil, fmt.Errorf("scan sector: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// MapObjects returns the blips to plot for one sector.
func (d *DB) MapObjects(ctx context.Context, sectorID int64) ([]MapObject, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT o.id, o.name, oc.kind, oc.name, o.x, o.y, o.status,
		       (o.owner_character_id IS NOT NULL)
		  FROM objects o
		  JOIN object_class oc ON oc.id = o.object_class_id
		 WHERE o.sector_id = $1
		 ORDER BY o.id`, sectorID)
	if err != nil {
		return nil, fmt.Errorf("map objects: %w", err)
	}
	defer rows.Close()
	out := []MapObject{}
	for rows.Next() {
		var m MapObject
		if err := rows.Scan(&m.ID, &m.Name, &m.Kind, &m.Class, &m.X, &m.Y, &m.Status, &m.Owned); err != nil {
			return nil, fmt.Errorf("scan map object: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// CredByEmail returns an account's stored credential for the placeholder admin
// login (bcrypt-verified by the api). ErrNotFound if no account has that email.
func (d *DB) CredByEmail(ctx context.Context, email string) (*AccountCred, error) {
	var c AccountCred
	err := d.pool.QueryRow(ctx,
		`SELECT id, email, password_hash FROM accounts WHERE email = $1`, email).
		Scan(&c.ID, &c.Email, &c.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("cred by email: %w", err)
	}
	return &c, nil
}

// rawOrEmpty guards against a NULL/empty JSONB column reaching the client as
// invalid JSON — it forwards a valid empty document instead.
func rawOrEmpty(b []byte) []byte {
	if len(b) == 0 {
		return []byte("null")
	}
	return b
}
