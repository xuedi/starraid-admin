package db

import (
	"encoding/json"
	"time"
)

// Summary is the Dashboard's at-a-glance DB counts (live server telemetry is a
// separate proxy path — see the API).
type Summary struct {
	Accounts   int          `json:"accounts"`
	Characters int          `json:"characters"`
	Objects    int          `json:"objects"`
	Sectors    int          `json:"sectors"`
	ByClass    []ClassCount `json:"by_class"`
}

// ClassCount is one row of the objects-per-class breakdown.
type ClassCount struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// UserRow is one account in the Users table.
type UserRow struct {
	ID          int64      `json:"id"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
	Characters  int        `json:"characters"`
}

// UserDetail is one account with its characters and their owned objects.
type UserDetail struct {
	ID          int64       `json:"id"`
	Email       string      `json:"email"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	LastLoginAt *time.Time  `json:"last_login_at"`
	Characters  []Character `json:"characters"`
}

// Character belongs to an account; it carries progression and owns objects.
type Character struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Prestige  int           `json:"prestige"`
	Faction   *string       `json:"faction"`
	CreatedAt time.Time     `json:"created_at"`
	Objects   []OwnedObject `json:"objects"`
}

// OwnedObject is the compact object summary shown under a character.
type OwnedObject struct {
	ID     int64   `json:"id"`
	Name   *string `json:"name"`
	Class  string  `json:"class"`
	Sector string  `json:"sector"`
}

// ObjectRow is one object in the Objects table.
type ObjectRow struct {
	ID        int64   `json:"id"`
	Name      *string `json:"name"`
	ClassKey  string  `json:"class_key"`
	ClassName string  `json:"class_name"`
	Kind      string  `json:"kind"`
	SectorID  int64   `json:"sector_id"`
	Sector    string  `json:"sector"`
	Owner     *string `json:"owner"`
	X         int64   `json:"x"`
	Y         int64   `json:"y"`
	Health    *int    `json:"health"`
	Shield    *int    `json:"shield"`
	Status    string  `json:"status"`
}

// ObjectDetail is the full object page: identity, class caps, position/sector,
// owner, live hull/shield, plus its fitting and cargo.
type ObjectDetail struct {
	ID               int64           `json:"id"`
	Name             *string         `json:"name"`
	Status           string          `json:"status"`
	X                int64           `json:"x"`
	Y                int64           `json:"y"`
	Health           *int            `json:"health"`
	Shield           *int            `json:"shield"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	Class            ClassInfo       `json:"class"`
	SectorID         int64           `json:"sector_id"`
	Sector           string          `json:"sector"`
	OwnerCharacterID *int64          `json:"owner_character_id"`
	Owner            *string         `json:"owner"`
	OwnerAccountID   *int64          `json:"owner_account_id"`
	Modules          []ModuleFitting `json:"modules"`
	Cargo            []CargoStack    `json:"cargo"`
}

// ClassInfo is the object's class (structure + slot capacity, from the catalog).
// Slots is forwarded as raw JSON ([{kind,size,count}]) — the frontend renders it.
type ClassInfo struct {
	Key             string          `json:"key"`
	Name            string          `json:"name"`
	Kind            string          `json:"kind"`
	SizeClass       string          `json:"size_class"`
	BaseMass        int64           `json:"base_mass"`
	BaseCargoVolume int64           `json:"base_cargo_volume"`
	Slots           json.RawMessage `json:"slots"`
}

// ModuleFitting is one fitted module (an object_modules row joined to its type).
// Params is forwarded as raw JSON (the module_types.params behaviour parameters).
type ModuleFitting struct {
	ID        int64           `json:"id"`
	Key       string          `json:"key"`
	Name      string          `json:"name"`
	SlotKind  string          `json:"slot_kind"`
	SlotIndex int             `json:"slot_index"`
	Quality   int             `json:"quality"`
	Status    string          `json:"status"`
	Params    json.RawMessage `json:"params"`
}

// CargoStack is one cargo stack (an object_items row joined to its item type).
type CargoStack struct {
	ID       int64  `json:"id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Quantity int64  `json:"quantity"`
}

// Sector is one row of the sector list (with its object count).
type Sector struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Objects   int       `json:"objects"`
}

// AccountCred is the internal credential row for the placeholder admin login
// (the password hash is never serialized to the client). See the api login
// handler — a "for now" front door, not the eventual admin auth (a parked TBD).
type AccountCred struct {
	ID           int64
	Email        string
	PasswordHash string
}

// MapObject is one blip on the sector map canvas.
type MapObject struct {
	ID     int64   `json:"id"`
	Name   *string `json:"name"`
	Kind   string  `json:"kind"`
	Class  string  `json:"class"`
	X      int64   `json:"x"`
	Y      int64   `json:"y"`
	Status string  `json:"status"`
	Owned  bool    `json:"owned"`
}
