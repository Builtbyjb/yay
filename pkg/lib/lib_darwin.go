//go:build darwin

package lib

import (
	"fmt"
	"slices"

	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/Builtbyjb/yay/pkg/lib/darwin"
)

func GetDatabase() (*core.Database, error) {

	dbPath, err := darwin.GetDatabasePath()
	if err != nil {
		return nil, err
	}

	db, err := core.NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Init(); err != nil {
		return nil, err
	}

	return db, nil
}

func Fetch() (*core.Database, []core.Setting, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, nil, err
	}

	dirs := darwin.AppDirectories
	settings, err := darwin.GetSettings(*db, dirs)
	if err != nil {
		return nil, nil, err
	}
	return db, settings, nil
}

func RawcodeToString(rawcode uint16) (string, error) {
	key, ok := darwin.RawToKeyDarwin[rawcode]
	if !ok {
		return "", fmt.Errorf("unknown rawcode: %d", rawcode)
	}
	return key, nil
}

func KeyEventListener(db *core.Database, onEvent func(KeyEvent)) {
	darwin.Listener(db, func(de darwin.KeyEvent) {
		if onEvent != nil {
			onEvent(KeyEvent{
				Keycode:   de.Keycode,
				Flags:     de.Flags,
				EventType: de.EventType,
			})
		}
	})
}

func VerifiedModifier(key string) bool {
	return slices.Contains(darwin.ModifiersMacos, key)
}
