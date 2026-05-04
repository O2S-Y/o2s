// Package plugin implements the `o2s plugin` command group — a lightweight
// extension mechanism: drop executables into the plugins directory and
// invoke them with `o2s plugin run <name> -- [...]`.
package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/config"
	"github.com/O2S-Y/o2s/internal/ui"
)

var safeName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\-.]*$`)

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"plugins", "pl"},
		Short:   "Discover and run drop-in plugin executables",
		Long: `Plugins are executables placed in your O2S plugins directory.

Environment passed to plugins:
  O2S_PLUGIN=1
  O2S_JSON=true|false   (mirrors global --json)
  O2S_THEME=...         (current theme name)
`,
	}
	c.AddCommand(pluginListCmd(), pluginRunCmd(), pluginPathCmd(), pluginInitCmd())
	return c
}

func pluginPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the plugins directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := config.PluginDir()
			if err != nil {
				return err
			}
			ui.Println(p)
			return nil
		},
	}
}

func pluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available plugin executables",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := config.PluginDir()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}

			ents, err := os.ReadDir(dir)
			if err != nil {
				return err
			}

			type pluginRow struct {
				Name string `json:"name"`
				Mode string `json:"mode"`
			}
			payload := struct {
				Dir     string      `json:"dir"`
				Plugins []pluginRow `json:"plugins"`
			}{Dir: dir}

			rows := [][]string{}
			for _, e := range ents {
				if e.IsDir() {
					continue
				}
				n := e.Name()
				if shouldSkipPluginListing(n) {
					continue
				}
				info, err := e.Info()
				if err != nil {
					continue
				}
				mode := info.Mode().String()
				rows = append(rows, []string{n, mode})
				payload.Plugins = append(payload.Plugins, pluginRow{Name: n, Mode: mode})
			}

			tbl := ui.Table{Headers: []string{"NAME", "MODE"}, Rows: rows}
			pretty := ui.Headline("Plugins") + "\n" + tbl.Render()
			return ui.Render(payload, pretty)
		},
	}
}

func shouldSkipPluginListing(name string) bool {
	ln := strings.ToLower(name)
	return strings.HasSuffix(ln, ".md") || strings.HasSuffix(ln, ".txt")
}

func pluginRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <name> [-- args...]",
		Short: "Execute a plugin by name (passes args after --)",
		Long: `Examples:

  o2s plugin run mytool -- --help
  o2s --json plugin run backup -- --dry-run

On Windows, .exe/.cmd/.bat/.ps1 are tried automatically if the bare name is missing.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if !safeName.MatchString(name) {
				return fmt.Errorf("invalid plugin name %q (allowed: letters, digits, ._-)", name)
			}
			rest := args[1:]

			dir, err := config.PluginDir()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}

			path, err := resolvePluginPath(dir, name)
			if err != nil {
				return err
			}

			cfg, _ := config.Load()
			env := os.Environ()
			env = append(env, "O2S_PLUGIN=1", "O2S_THEME="+cfg.Theme)
			if ui.JSON() {
				env = append(env, "O2S_JSON=true")
			} else {
				env = append(env, "O2S_JSON=false")
			}

			c := exec.Command(path, rest...)
			c.Env = env
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if wd, err := os.Getwd(); err == nil {
				c.Dir = wd
			}
			return c.Run()
		},
	}
}

func pluginInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create the plugins directory and a starter README",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := config.PluginDir()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			readme := filepath.Join(dir, "README.txt")
			content := `O2S plugins directory
=====================

Drop executables here and run them with:

  o2s plugin list
  o2s plugin run <name> -- [args...]

Plugins receive:

  O2S_PLUGIN=1
  O2S_JSON=true|false
  O2S_THEME=<name>
`
			_ = os.WriteFile(readme, []byte(content), 0o644)
			ui.Success("plugins directory ready at " + dir)
			return nil
		},
	}
}

func resolvePluginPath(dir, name string) (string, error) {
	candidates := []string{filepath.Join(dir, name)}
	if runtime.GOOS == "windows" {
		candidates = append(candidates,
			filepath.Join(dir, name+".exe"),
			filepath.Join(dir, name+".cmd"),
			filepath.Join(dir, name+".bat"),
			filepath.Join(dir, name+".ps1"),
		)
	}
	for _, p := range candidates {
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			continue
		}
		return filepath.Clean(p), nil
	}
	return "", fmt.Errorf("plugin %q not found in %s", name, dir)
}
