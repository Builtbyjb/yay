package core

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	conn *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &Database{conn: db}, nil
}

func (d *Database) Init() error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		bin_name TEXT NOT NULL,
		path TEXT NOT NULL,
		hotkey TEXT UNIQUE,
		mode TEXT CHECK(mode IN ('default', 'desktop')),
		enabled BOOLEAN
	);`

	_, err := d.conn.Exec(createTableQuery)
	if err != nil {
		return err
	}

	createHotkeyIndex := `CREATE INDEX IF NOT EXISTS idx_hotkey ON settings (hotkey)`
	_, err = d.conn.Exec(createHotkeyIndex)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) Close() error {
	return d.conn.Close()
}

func (d *Database) Insert(name string, path string, binName string, hotkey sql.NullString, mode string, enabled bool) error {
	query := "INSERT INTO settings (name, path, bin_name, hotkey, mode, enabled) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := d.conn.Exec(query, name, path, binName, hotkey, mode, enabled)
	return err
}

func (d *Database) UpdateEnabled(id int, enabled bool) error {
	query := "UPDATE settings SET enabled = ? WHERE id = ? "
	_, err := d.conn.Exec(query, enabled, id)
	return err
}

func (d *Database) UpdateMode(id int, mode string) error {
	query := "update settings set mode = ? where id = ? "
	_, err := d.conn.Exec(query, mode, id)
	return err
}

func (d *Database) FindByHotkey(hotkey string) (*Setting, error) {
	query := "SELECT * FROM settings WHERE hotkey = ?"
	row := d.conn.QueryRow(query, hotkey)

	var s Setting
	if err := row.Scan(&s.Id, &s.Name, &s.BinName, &s.Path, &s.HotKey, &s.Mode, &s.Enabled); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (d *Database) UpdateHotkey(id int, hotkey sql.NullString) error {
	query := "UPDATE settings SET hotkey = ? WHERE id = ? "
	_, err := d.conn.Exec(query, hotkey, id)
	return err
}

func (d *Database) ClearHotkey(id int) error {
	query := "UPDATE settings SET hotkey = ? WHERE id = ?"
	_, err := d.conn.Exec(query, nil, id)
	return err
}

func (d *Database) Refresh(apps []App) ([]Setting, error) {
	// Clear settings
	_, err := d.conn.Exec("DELETE FROM settings")
	if err != nil {
		return nil, err
	}

	// Add apps in apps list but not in the database
	for _, app := range apps {
		_, err := d.conn.Exec(
			"INSERT INTO settings (name, bin_name, path, hotkey, mode, enabled) VALUES (?, ?, ?, ?, ?, ?)",
			app.Name, app.BinName, app.Path, sql.NullString{String: "", Valid: false}, "default", true,
		)
		if err != nil {
			return nil, err
		}
	}

	// Return the updated settings list
	settings, err := d.GetAllSettings()
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (d *Database) GetAllSettings() ([]Setting, error) {
	updatedRows, err := d.conn.Query("SELECT * FROM settings ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer updatedRows.Close()

	var settings []Setting
	for updatedRows.Next() {
		var s Setting
		if err := updatedRows.Scan(&s.Id, &s.Name, &s.BinName, &s.Path, &s.HotKey, &s.Mode, &s.Enabled); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	if err := updatedRows.Err(); err != nil {
		return nil, err
	}

	return settings, nil
}
