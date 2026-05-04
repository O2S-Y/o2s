package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func findCmd() *cobra.Command {
	var (
		root  string
		limit int
		all   bool
	)
	c := &cobra.Command{
		Use:   "find <query>",
		Short: "Fuzzy-find files under a directory (skips noisy folders)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			start := root
			if start == "" {
				start = "."
			}
			abs, err := filepath.Abs(start)
			if err != nil {
				return err
			}
			candidates, err := collectFiles(abs, all)
			if err != nil {
				return err
			}
			matches := fuzzy.Find(query, candidates)
			t := ui.T()
			ui.Println(ui.Headline(fmt.Sprintf("%d match(es) for %q", len(matches), query)))
			max := limit
			if max == 0 || max > len(matches) {
				max = len(matches)
			}
			for i := 0; i < max; i++ {
				m := matches[i]
				orig := m.Str
				fmt.Fprintln(ui.Out, t.Key.Render(fmt.Sprintf("%2d.", i+1))+" "+highlight(orig, m.MatchedIndexes))
			}
			return nil
		},
	}
	c.Flags().StringVar(&root, "root", "", "directory to search under (default: cwd)")
	c.Flags().IntVarP(&limit, "limit", "n", 30, "max results to show (0 = all)")
	c.Flags().BoolVarP(&all, "all", "a", false, "include hidden / ignored folders")
	return c
}

func collectFiles(root string, all bool) ([]string, error) {
	out := make([]string, 0, 1024)
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		base := filepath.Base(p)
		if !all {
			if strings.HasPrefix(base, ".") && p != root {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if _, ok := skip[base]; ok && d.IsDir() {
				return filepath.SkipDir
			}
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, p)
			out = append(out, rel)
		}
		return nil
	})
	return out, err
}

// highlight underlines/bolds the matched runes for a snappier search feel.
func highlight(s string, idx []int) string {
	if len(idx) == 0 {
		return s
	}
	t := ui.T()
	in := map[int]struct{}{}
	for _, i := range idx {
		in[i] = struct{}{}
	}
	var b strings.Builder
	for i, r := range s {
		if _, ok := in[i]; ok {
			b.WriteString(t.Key.Render(string(r)))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
