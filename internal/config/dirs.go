package config

import (
	"os"
	"path/filepath"
)

// PluginDir returns the directory where drop-in plugin executables live.
// Default: <config>/o2s/plugins (override with O2S_PLUGIN_DIR).
func PluginDir() (string, error) {
	if v := os.Getenv("O2S_PLUGIN_DIR"); v != "" {
		return v, nil
	}
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "plugins"), nil
}
