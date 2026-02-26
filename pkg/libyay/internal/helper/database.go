package helper

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	conn *sql.DB
}

func NewDatabase(dbPath string) *Database {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	return &Database{conn: db}
}

func (d *Database) Init() {
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
		log.Fatal(err)
	}
}

func (d *Database) UpdateHokey(setting Setting) {

}

func (d *Database) UpdateMode(setting Setting) {

}

func (d *Database) UpdateEnabled(setting Setting) {

}

func (d *Database) Refresh(apps []App) []Setting {
	// NOTE: Return settings id
	// TODO: Break it up into private methods
	// Build a map of app paths for quick lookup
	appMap := make(map[string]App)
	for _, app := range apps {
		appMap[app.Path] = app
	}

	// Fetch all existing settings from the database
	rows, err := d.conn.Query("SELECT name, path, icon_path, hotkey, mode, enabled FROM settings")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	existingPaths := make(map[string]struct{})
	for rows.Next() {
		var s Setting
		if err := rows.Scan(&s.Name, &s.Path, &s.IconPath, &s.HotKey, &s.Mode, &s.Enabled); err != nil {
			log.Fatal(err)
		}
		existingPaths[s.Path] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Add apps in apps list but not in the database
	// TODO: Use the save method here
	for _, app := range apps {
		if _, exists := existingPaths[app.Path]; !exists {
			_, err := d.conn.Exec(
				"INSERT INTO settings (name, path, icon_path, hotkey, mode, enabled) VALUES (?, ?, ?, ?, ?, ?)",
				app.Name, app.Path, app.IconPath, "", "default", true,
			)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Remove apps in the database but not in the apps list
	for path := range existingPaths {
		if _, exists := appMap[path]; !exists {
			_, err := d.conn.Exec("DELETE FROM settings WHERE path = ?", path)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Return the updated settings list
	updatedRows, err := d.conn.Query("SELECT name, path, icon_path, hotkey, mode, enabled FROM settings")
	if err != nil {
		log.Fatal(err)
	}
	defer updatedRows.Close()

	var settings []Setting
	for updatedRows.Next() {
		var s Setting
		if err := updatedRows.Scan(&s.Name, &s.Path, &s.IconPath, &s.HotKey, &s.Mode, &s.Enabled); err != nil {
			log.Fatal(err)
		}
		settings = append(settings, s)
	}
	if err := updatedRows.Err(); err != nil {
		log.Fatal(err)
	}

	return settings
}

func (d *Database) Close() error {
	return d.conn.Close()
}
