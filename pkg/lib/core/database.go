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
	// Create settings table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT NOT NULL,
		icon_path TEXT NOT NULL,
		hotkey TEXT,
		mode TEXT,
		enabled BOOLEAN
	);
	`
	_, err := d.conn.Exec(createTableQuery)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) Close() error {
	return d.conn.Close()
}

func (d *Database) UpdateHokey(id int, hotkey string) error {
	query := "UPDATE settings SET hotkey = ? WHERE id = ?"
	_, err := d.conn.Exec(query, hotkey, id)
	return err
}

func (d *Database) UpdateMode(id int, mode string) error {
	query := "UPDATE settings SET mode = ? WHERE id = ?"
	_, err := d.conn.Exec(query, mode, id)
	return err
}

func (d *Database) UpdateEnabled(id int, enabled bool) error {
	query := "UPDATE settings SET enabled = ? WHERE id = ?"
	_, err := d.conn.Exec(query, enabled, id)
	return err
}

func (d *Database) Refresh(apps []App) ([]Setting, error) {
	// Build a map of app paths for quick lookup
	appMap := make(map[string]App)
	for _, app := range apps {
		appMap[app.Path] = app
	}

	// Fetch all existing settings from the database
	existingPaths, err := d.getExistingPaths()
	if err != nil {
		return nil, err
	}

	// Add apps in apps list but not in the database
	for _, app := range apps {
		if _, exists := existingPaths[app.Path]; !exists {
			_, err := d.conn.Exec(
				"INSERT INTO settings (name, path, icon_path, hotkey, mode, enabled) VALUES (?, ?, ?, ?, ?, ?)",
				app.Name, app.Path, app.IconPath, "", "default", true,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	// Remove apps in the database but not in the apps list
	for path := range existingPaths {
		if _, exists := appMap[path]; !exists {
			_, err := d.conn.Exec("DELETE FROM settings WHERE path = ?", path)
			if err != nil {
				return nil, err
			}
		}
	}

	// Return the updated settings list
	settings, err := d.getUpdatedSettings()
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (d *Database) getExistingPaths() (map[string]struct{}, error) {
	rows, err := d.conn.Query("SELECT * FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existingPaths := make(map[string]struct{})
	for rows.Next() {
		var s Setting
		if err := rows.Scan(&s.Id, &s.Name, &s.Path, &s.IconPath, &s.HotKey, &s.Mode, &s.Enabled); err != nil {
			return nil, err
		}
		existingPaths[s.Path] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return existingPaths, nil
}

func (d *Database) getUpdatedSettings() ([]Setting, error) {
	updatedRows, err := d.conn.Query("SELECT * FROM settings")
	if err != nil {
		return nil, err
	}
	defer updatedRows.Close()

	var settings []Setting
	for updatedRows.Next() {
		var s Setting
		if err := updatedRows.Scan(&s.Id, &s.Name, &s.Path, &s.IconPath, &s.HotKey, &s.Mode, &s.Enabled); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	if err := updatedRows.Err(); err != nil {
		return nil, err
	}

	return settings, nil
}
