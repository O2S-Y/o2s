package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

// dedupCmd finds duplicate files by hashing their contents.
// Strategy: bucket by size first (cheap), then sha256 inside each bucket.
func dedupCmd() *cobra.Command {
	var (
		minSize int64
	)
	c := &cobra.Command{
		Use:   "dedup [path]",
		Short: "Find duplicate files (size + sha256)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			abs, _ := filepath.Abs(root)
			ui.Println(ui.Headline("Scanning " + abs))

			bySize := map[int64][]string{}
			err := filepath.WalkDir(abs, func(p string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				info, err := d.Info()
				if err != nil {
					return nil
				}
				if info.Size() < minSize {
					return nil
				}
				bySize[info.Size()] = append(bySize[info.Size()], p)
				return nil
			})
			if err != nil {
				return err
			}

			byHash := map[string][]string{}
			for size, paths := range bySize {
				if len(paths) < 2 {
					continue
				}
				for _, p := range paths {
					h, err := quickHash(p)
					if err != nil {
						continue
					}
					key := fmt.Sprintf("%d:%s", size, h)
					byHash[key] = append(byHash[key], p)
				}
			}

			type group struct {
				Size    int64    `json:"size"`
				Hash    string   `json:"hash"`
				Members []string `json:"members"`
			}
			groups := make([]group, 0)
			for k, v := range byHash {
				if len(v) < 2 {
					continue
				}
				var size int64
				var h string
				fmt.Sscanf(k, "%d:%s", &size, &h)
				groups = append(groups, group{Size: size, Hash: h, Members: v})
			}
			sort.Slice(groups, func(i, j int) bool { return groups[i].Size > groups[j].Size })

			t := ui.T()
			savings := int64(0)
			for _, g := range groups {
				ui.Println("")
				ui.Println(t.Title.Render(fmt.Sprintf("× %d copies — %s — %s", len(g.Members), humanize.IBytes(uint64(g.Size)), g.Hash[:12])))
				for _, m := range g.Members {
					ui.Println("  " + t.Muted.Render("•") + " " + m)
				}
				savings += g.Size * int64(len(g.Members)-1)
			}
			ui.Println("")
			ui.Println(t.Success.Render(fmt.Sprintf("potential savings: %s across %d duplicate set(s)", humanize.IBytes(uint64(savings)), len(groups))))
			return nil
		},
	}
	c.Flags().Int64VarP(&minSize, "min-size", "m", 1024, "minimum file size in bytes (skip tiny files)")
	return c
}

func quickHash(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
