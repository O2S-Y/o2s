package dev

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func gitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "git",
		Short: "Git helpers (multi-repo status, sync)",
	}
	c.AddCommand(gitStatusCmd(), gitSyncCmd())
	return c
}

// gitStatusCmd walks subdirectories that look like git repos and renders a
// colourised status overview — handy when you have a workspace folder of repos.
func gitStatusCmd() *cobra.Command {
	var depth int
	c := &cobra.Command{
		Use:   "status [path]",
		Short: "Multi-repo git status (recursive scan)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			abs, _ := filepath.Abs(root)
			repos := findRepos(abs, depth)

			tbl := ui.Table{Headers: []string{"REPO", "BRANCH", "AHEAD/BEHIND", "DIRTY", "STATUS"}}
			t := ui.T()
			for _, r := range repos {
				branch := gitOutput(r, "rev-parse", "--abbrev-ref", "HEAD")
				ahead := gitOutput(r, "rev-list", "--count", "@{u}..HEAD")
				behind := gitOutput(r, "rev-list", "--count", "HEAD..@{u}")
				dirty := strings.TrimSpace(gitOutput(r, "status", "--porcelain"))
				dirtyMark := t.Success.Render("clean")
				if dirty != "" {
					dirtyMark = t.Warning.Render("dirty")
				}
				ab := "-"
				if ahead != "" || behind != "" {
					ab = fmt.Sprintf("%s↑ %s↓", strings.TrimSpace(ahead), strings.TrimSpace(behind))
				}
				rel, _ := filepath.Rel(abs, r)
				tbl.Rows = append(tbl.Rows, []string{
					rel, strings.TrimSpace(branch), ab, dirtyMark, ui.Status("ok", "ok"),
				})
			}
			ui.Println(ui.Headline(fmt.Sprintf("Git status — %d repo(s)", len(repos))))
			ui.Println(tbl.Render())
			return nil
		},
	}
	c.Flags().IntVarP(&depth, "depth", "d", 3, "max recursion depth")
	return c
}

func gitSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [path]",
		Short: "Pull (and optionally push) for the current repo",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			steps := []struct {
				label string
				args  []string
			}{
				{"fetch", []string{"fetch", "--all", "--prune"}},
				{"pull --rebase", []string{"pull", "--rebase"}},
			}
			for _, s := range steps {
				ui.Info("git " + s.label)
				cmd := exec.Command("git", s.args...)
				cmd.Dir = root
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}
			}
			ui.Success("synced")
			return nil
		},
	}
}

func findRepos(root string, maxDepth int) []string {
	var repos []string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		depth := strings.Count(rel, string(filepath.Separator))
		if depth > maxDepth {
			return filepath.SkipDir
		}
		if d.IsDir() && d.Name() == ".git" {
			repos = append(repos, filepath.Dir(p))
			return filepath.SkipDir
		}
		return nil
	})
	return repos
}

func gitOutput(dir string, args ...string) string {
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
