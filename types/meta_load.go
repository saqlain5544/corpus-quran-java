package types

import (
	"encoding/json"
	"os"
)

// loadMetaJSON reads a meta.json file from disk into a Meta struct.
// The Meta is small (~3 KB) and JSON-encoded for easy inspection by
// humans during development.
func loadMetaJSON(path string) (*Meta, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Meta
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
