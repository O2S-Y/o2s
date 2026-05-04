package dev

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

// initCmd scaffolds a new project of a chosen template into a target directory.
// Templates are intentionally minimal — just enough to start hacking.
func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [name]",
		Short: "Scaffold a new project (Huh-driven wizard)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			tmpl := "node"
			gitInit := true

			if !ui.IsTTY() {
				return fmt.Errorf("`o2s dev init` needs an interactive terminal")
			}
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Project name").Value(&name).Placeholder("my-app"),
					huh.NewSelect[string]().Title("Template").Options(
						huh.NewOption("Node + TypeScript", "node"),
						huh.NewOption("Go module", "go"),
						huh.NewOption("Python (uv-style)", "python"),
						huh.NewOption("Static site (HTML/CSS/JS)", "static"),
					).Value(&tmpl),
					huh.NewConfirm().Title("Run `git init`?").Value(&gitInit).Affirmative("Yes").Negative("No"),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
			if name == "" {
				return fmt.Errorf("project name is required")
			}
			dest, err := filepath.Abs(name)
			if err != nil {
				return err
			}
			if _, err := os.Stat(dest); err == nil {
				return fmt.Errorf("directory %s already exists", dest)
			}
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
			files := templateFor(tmpl, name)
			for path, content := range files {
				p := filepath.Join(dest, path)
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
					return err
				}
			}
			if gitInit {
				c := exec.Command("git", "init")
				c.Dir = dest
				if err := c.Run(); err != nil {
					ui.Warn("git init failed: " + err.Error())
				} else {
					ui.Info("git initialised")
				}
			}
			ui.Success("scaffolded " + dest)
			ui.Info("next: cd " + name)
			return nil
		},
	}
}

func templateFor(tmpl, name string) map[string]string {
	switch tmpl {
	case "go":
		return map[string]string{
			"go.mod": fmt.Sprintf("module %s\n\ngo 1.22\n", strings.ToLower(name)),
			"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello from ` + name + `")
}
`,
			"README.md": "# " + name + "\n\nBootstrapped with `o2s dev init`.\n",
			".gitignore": "/" + name + "\n*.exe\n",
		}
	case "python":
		return map[string]string{
			"pyproject.toml": fmt.Sprintf("[project]\nname = \"%s\"\nversion = \"0.0.1\"\nrequires-python = \">=3.10\"\n", name),
			"src/main.py":    "def main():\n    print(\"Hello from " + name + "\")\n\n\nif __name__ == \"__main__\":\n    main()\n",
			"README.md":      "# " + name + "\n",
			".gitignore":     ".venv/\n__pycache__/\n*.pyc\n",
		}
	case "static":
		return map[string]string{
			"index.html": "<!doctype html>\n<html>\n  <head><title>" + name + "</title></head>\n  <body><h1>" + name + "</h1></body>\n</html>\n",
			"style.css":  "body { font-family: system-ui, sans-serif; }\n",
		}
	default: // node
		return map[string]string{
			"package.json": `{
  "name": "` + strings.ToLower(name) + `",
  "version": "0.0.1",
  "type": "module",
  "scripts": { "dev": "node src/index.js" }
}
`,
			"src/index.js": "console.log(\"Hello from " + name + "\");\n",
			"README.md":    "# " + name + "\n\nBootstrapped with `o2s dev init`.\n",
			".gitignore":   "node_modules/\ndist/\n",
		}
	}
}
