// Package config loads/persists user preferences (~/.config/o2s/config.yaml).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config is the strongly-typed shape of the YAML config.
type Config struct {
	Name     string `mapstructure:"name"      yaml:"name"`
	Theme    string `mapstructure:"theme"     yaml:"theme"`
	Editor   string `mapstructure:"editor"    yaml:"editor"`
	NoColor  bool   `mapstructure:"no_color"  yaml:"no_color"`
	Onboard  bool   `mapstructure:"onboarded" yaml:"onboarded"`
	JSONMode bool   `mapstructure:"json"      yaml:"json"`
}

// Default returns sensible defaults for a fresh install.
func Default() Config {
	return Config{
		Name:    "",
		Theme:   "default",
		Editor:  defaultEditor(),
		NoColor: false,
		Onboard: false,
	}
}

// Dir returns the per-user O2S config directory.
// Uses %APPDATA%\o2s on Windows and $XDG_CONFIG_HOME/o2s (or ~/.config/o2s) on Unix.
func Dir() (string, error) {
	if v := os.Getenv("O2S_CONFIG_DIR"); v != "" {
		return v, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "o2s"), nil
}

// Path returns the absolute path of the config file.
func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.yaml"), nil
}

// DataDir returns a directory for persistent data (todos, notes, db).
func DataDir() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "data"), nil
}

// Load reads the config file (creating defaults if missing) and returns the parsed value.
func Load() (Config, error) {
	c := Default()

	dir, err := Dir()
	if err != nil {
		return c, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return c, err
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.SetEnvPrefix("O2S")
	v.AutomaticEnv()

	v.SetDefault("theme", c.Theme)
	v.SetDefault("editor", c.Editor)
	v.SetDefault("no_color", c.NoColor)
	v.SetDefault("onboarded", c.Onboard)

	if err := v.ReadInConfig(); err != nil {
		var nfe viper.ConfigFileNotFoundError
		if !asViperNotFound(err, &nfe) {
			return c, fmt.Errorf("read config: %w", err)
		}
	}
	if err := v.Unmarshal(&c); err != nil {
		return c, fmt.Errorf("parse config: %w", err)
	}
	return c, nil
}

// Save writes the given config back to disk.
func Save(c Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("name", c.Name)
	v.Set("theme", c.Theme)
	v.Set("editor", c.Editor)
	v.Set("no_color", c.NoColor)
	v.Set("onboarded", c.Onboard)
	v.Set("json", c.JSONMode)
	return v.WriteConfigAs(p)
}

// asViperNotFound is a tiny helper so the caller above stays readable.
func asViperNotFound(err error, into *viper.ConfigFileNotFoundError) bool {
	if e, ok := err.(viper.ConfigFileNotFoundError); ok {
		*into = e
		return true
	}
	return false
}

func defaultEditor() string {
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	if isWindows() {
		return "notepad"
	}
	return "vi"
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
