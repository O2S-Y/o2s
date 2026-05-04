package dev

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

// runCmd discovers scripts in well-known files (package.json, Makefile,
// Taskfile.yml) and lets the user pick one with a styled prompt.
func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [name]",
		Short: "Run a script discovered in package.json / Makefile / Taskfile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scripts, err := discoverScripts()
			if err != nil {
				return err
			}
			if len(scripts) == 0 {
				return fmt.Errorf("no scripts found in this directory")
			}
			labels := make([]string, 0, len(scripts))
			for _, s := range scripts {
				labels = append(labels, fmt.Sprintf("[%s] %s — %s", s.Source, s.Name, s.Cmd))
			}
			var picked string
			if len(args) == 1 {
				for i, s := range scripts {
					if s.Name == args[0] {
						picked = labels[i]
						break
					}
				}
			}
			if picked == "" {
				p, err := ui.Select("Pick a script", labels)
				if err != nil {
					return err
				}
				picked = p
			}
			var chosen *script
			for i, l := range labels {
				if l == picked {
					chosen = &scripts[i]
					break
				}
			}
			if chosen == nil {
				return fmt.Errorf("nothing selected")
			}
			return execScript(*chosen)
		},
	}
}

type script struct {
	Source string
	Name   string
	Cmd    string
}

func discoverScripts() ([]script, error) {
	var out []script
	if data, err := os.ReadFile("package.json"); err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if err := json.Unmarshal(data, &pkg); err == nil {
			for name, cmd := range pkg.Scripts {
				out = append(out, script{Source: "npm", Name: name, Cmd: cmd})
			}
		}
	}
	if data, err := os.ReadFile("Makefile"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") || strings.TrimSpace(line) == "" {
				continue
			}
			if i := strings.Index(line, ":"); i > 0 {
				name := strings.TrimSpace(line[:i])
				if name != "" && !strings.ContainsAny(name, " =$") {
					out = append(out, script{Source: "make", Name: name, Cmd: "make " + name})
				}
			}
		}
	}
	for _, fn := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		if data, err := os.ReadFile(fn); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
					continue
				}
				if strings.HasPrefix(line, "tasks:") {
					continue
				}
				if i := strings.Index(line, ":"); i > 0 && !strings.HasPrefix(line, " ") {
					name := strings.TrimSpace(line[:i])
					if filepath.Ext(name) == "" && name != "" {
						out = append(out, script{Source: "task", Name: name, Cmd: "task " + name})
					}
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func execScript(s script) error {
	t := ui.T()
	ui.Println(t.Title.Render("▶ ") + t.Body.Render(s.Cmd))
	var c *exec.Cmd
	switch s.Source {
	case "npm":
		c = exec.Command("npm", "run", s.Name)
	case "make":
		c = exec.Command("make", s.Name)
	case "task":
		c = exec.Command("task", s.Name)
	default:
		c = exec.Command("sh", "-c", s.Cmd)
	}
	c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
	return c.Run()
}
