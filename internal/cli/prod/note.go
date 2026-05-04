package prod

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/config"
	"github.com/O2S-Y/o2s/internal/ui"
)

func noteCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "note",
		Aliases: []string{"n"},
		Short:   "Markdown notes (rendered with Glamour)",
	}
	c.AddCommand(noteNewCmd(), noteListCmd(), noteOpenCmd(), noteShowCmd())
	return c
}

func notesDir() (string, error) {
	d, err := config.DataDir()
	if err != nil {
		return "", err
	}
	nd := filepath.Join(d, "notes")
	return nd, os.MkdirAll(nd, 0o755)
}

func noteNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new <title...>",
		Short: "Create a new markdown note and open it in your editor",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			dir, err := notesDir()
			if err != nil {
				return err
			}
			fname := time.Now().Format("2006-01-02_") + slugify(title) + ".md"
			path := filepath.Join(dir, fname)
			body := fmt.Sprintf("# %s\n\n_created %s_\n\n", title, time.Now().Format(time.RFC1123))
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				return err
			}
			ui.Success("created " + path)

			cfg, _ := config.Load()
			editor := cfg.Editor
			if editor == "" {
				editor = "notepad"
			}
			c := exec.Command(editor, path)
			c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
			_ = c.Run()
			return nil
		},
	}
}

func noteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List markdown notes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := notesDir()
			if err != nil {
				return err
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				return err
			}
			notes := entries[:0]
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					notes = append(notes, e)
				}
			}
			sort.Slice(notes, func(i, j int) bool { return notes[i].Name() > notes[j].Name() })

			th := ui.T()
			if len(notes) == 0 {
				ui.Println(th.Muted.Render("no notes yet — try `o2s prod note new \"first note\"`"))
				return nil
			}
			tbl := ui.Table{Headers: []string{"FILENAME", "MODIFIED"}}
			for _, n := range notes {
				info, _ := n.Info()
				mod := "-"
				if info != nil {
					mod = info.ModTime().Format("2006-01-02 15:04")
				}
				tbl.Rows = append(tbl.Rows, []string{n.Name(), mod})
			}
			pretty := ui.Headline(fmt.Sprintf("Notes (%d)", len(notes))) + "\n" + tbl.Render()
			ui.Println(pretty)
			return nil
		},
	}
}

func noteOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <name>",
		Short: "Open a note in your editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveNote(args[0])
			if err != nil {
				return err
			}
			cfg, _ := config.Load()
			editor := cfg.Editor
			if editor == "" {
				editor = "notepad"
			}
			c := exec.Command(editor, path)
			c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
			return c.Run()
		},
	}
}

func noteShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Render a note in the terminal (Glamour)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveNote(args[0])
			if err != nil {
				return err
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			r, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(100))
			if err != nil {
				return err
			}
			out, err := r.Render(string(raw))
			if err != nil {
				return err
			}
			ui.Println(out)
			return nil
		},
	}
}

func resolveNote(name string) (string, error) {
	dir, err := notesDir()
	if err != nil {
		return "", err
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() == name || strings.Contains(strings.ToLower(e.Name()), strings.ToLower(name)) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no note matched %q", name)
}

func slugify(s string) string {
	out := strings.Builder{}
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			out.WriteRune('-')
		}
	}
	res := out.String()
	if res == "" {
		return "note"
	}
	return res
}
