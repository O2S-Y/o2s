package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func treeCmd() *cobra.Command {
	var (
		maxDepth int
		showAll  bool
		dirsOnly bool
	)
	c := &cobra.Command{
		Use:   "tree [path]",
		Short: "Pretty directory tree (skips .git, node_modules, etc by default)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			abs, err := filepath.Abs(root)
			if err != nil {
				return err
			}
			t := ui.T()
			ui.Println(t.Title.Render(abs))
			return walkTree(abs, "", 0, maxDepth, showAll, dirsOnly)
		},
	}
	c.Flags().IntVarP(&maxDepth, "depth", "L", 4, "max recursion depth (0 = unlimited)")
	c.Flags().BoolVarP(&showAll, "all", "a", false, "include hidden / ignored entries")
	c.Flags().BoolVarP(&dirsOnly, "dirs", "d", false, "show directories only")
	return c
}

var skip = map[string]struct{}{
	".git": {}, "node_modules": {}, ".cache": {}, ".idea": {}, ".vscode": {},
	"dist": {}, "build": {}, "target": {}, ".next": {}, ".venv": {},
	"__pycache__": {}, ".DS_Store": {},
}

func walkTree(dir, prefix string, depth, maxDepth int, all, dirsOnly bool) error {
	if maxDepth > 0 && depth >= maxDepth {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	filtered := entries[:0]
	for _, e := range entries {
		name := e.Name()
		if !all {
			if strings.HasPrefix(name, ".") {
				continue
			}
			if _, ok := skip[name]; ok {
				continue
			}
		}
		if dirsOnly && !e.IsDir() {
			continue
		}
		filtered = append(filtered, e)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].IsDir() != filtered[j].IsDir() {
			return filtered[i].IsDir()
		}
		return filtered[i].Name() < filtered[j].Name()
	})

	t := ui.T()
	for i, e := range filtered {
		isLast := i == len(filtered)-1
		branch := "├── "
		nextPrefix := prefix + "│   "
		if isLast {
			branch = "└── "
			nextPrefix = prefix + "    "
		}
		name := e.Name()
		styled := name
		if e.IsDir() {
			styled = t.Title.Render(name + "/")
		} else if strings.HasPrefix(name, ".") {
			styled = t.Muted.Render(name)
		}
		fmt.Fprintln(ui.Out, t.Muted.Render(prefix+branch)+styled)
		if e.IsDir() {
			if err := walkTree(filepath.Join(dir, name), nextPrefix, depth+1, maxDepth, all, dirsOnly); err != nil {
				continue
			}
		}
	}
	return nil
}
