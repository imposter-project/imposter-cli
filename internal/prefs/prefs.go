package prefs

import (
	"encoding/json"
	"fmt"
	"gatehill.io/imposter/internal/library"
	"os"
	"path/filepath"
)

type Preferences interface {
	// ReadPropertyString attempts to read the property with the given key
	// from the prefs store and then attempts to type assert any non-nil value
	// as a string. If the store does not exist, or if the store cannot be parsed,
	// or the property does not exist, the empty string is returned.
	ReadPropertyString(key string) (string, error)

	// ReadPropertyInt attempts to read the property with the given key
	// from the prefs store and then attempts to type assert any non-nil value
	// as an int. If the store does not exist, or if the store cannot be parsed,
	// or the property does not exist, 0 is returned.
	ReadPropertyInt(key string) (int, error)

	// WriteProperty persists a key-value pair to the prefs store.
	// If the store cannot be loaded, the write will fail and an error
	// will be logged.
	WriteProperty(key string, value interface{}) error
}

type Prefs struct {
	fileName string
}

func Load(fileName string) Prefs {
	return Prefs{fileName: fileName}
}

func (p Prefs) loadState() (map[string]interface{}, error) {
	prefsFile, err := p.ensurePrefsFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get prefs file: %s", err)
	}
	file, err := os.ReadFile(prefsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to read prefs file: %s: %s", prefsFile, err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall prefs file: %s: %s", prefsFile, err)
	}
	return data, nil
}

func (p Prefs) ensurePrefsFile() (string, error) {
	return p.ensurePrefsFileWithName(p.fileName)
}

func (p Prefs) ensurePrefsFileWithName(fileName string) (string, error) {
	prefsDir, err := p.ensurePrefsDir()
	if err != nil {
		return "", fmt.Errorf("failed to ensure prefs dir exists: %s", err)
	}
	return filepath.Join(prefsDir, fileName), nil
}

func (p Prefs) ensurePrefsDir() (string, error) {
	return library.EnsureDirUsingConfig("prefs.dir", ".imposter")
}

// ReadPropertyString attempts to read the property with the given key
// from the prefs store and then attempts to type assert any non-nil value
// as a string. If the store does not exist, or if the store cannot be parsed,
// or the property does not exist, the empty string is returned.
func (p Prefs) ReadPropertyString(key string) (string, error) {
	value, err := p.readProperty(key)
	if err != nil || value == nil {
		return "", err
	}
	return value.(string), nil
}

// ReadPropertyInt attempts to read the property with the given key
// from the prefs store and then attempts to type assert any non-nil value
// as an int. If the store does not exist, or if the store cannot be parsed,
// or the property does not exist, 0 is returned.
func (p Prefs) ReadPropertyInt(key string) (int, error) {
	value, err := p.readProperty(key)
	if err != nil || value == nil {
		return 0, err
	}
	return int(value.(float64)), nil
}

func (p Prefs) readProperty(key string) (interface{}, error) {
	state, err := p.loadState()
	if err != nil {
		return "", err
	}
	return state[key], nil
}

// WriteProperty persists a key-value pair to the prefs store.
// If the store cannot be loaded, the write will fail and an error
// will be logged.
func (p Prefs) WriteProperty(key string, value interface{}) error {
	state, err := p.loadState()
	if err != nil {
		return fmt.Errorf("failed to read existing prefs: %s", err)
	}
	state[key] = value
	j, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshall prefs: %s", err)
	}
	prefsFile, err := p.ensurePrefsFile()
	if err != nil {
		return fmt.Errorf("failed to get prefs file: %s", err)
	}
	err = os.WriteFile(prefsFile, j, 0644)
	if err != nil {
		return fmt.Errorf("failed to write prefs file: %s", err)
	}
	return nil
}
