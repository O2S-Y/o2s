package prod

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/store"
	"github.com/O2S-Y/o2s/internal/ui"
)

type Snippet struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func snippetCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "snippet",
		Aliases: []string{"snippets", "snip"},
		Short:   "Code/text snippets with copy-to-clipboard",
	}
	c.AddCommand(snippetAddCmd(), snippetListCmd(), snippetGetCmd(), snippetRmCmd(), snippetCopyCmd())
	return c
}

func snippetAddCmd() *cobra.Command {
	var (
		body string
		from string
		tags string
	)
	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Create or replace a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := normalizeSnippetName(args[0])
			content := strings.TrimSpace(body)
			if from != "" {
				raw, err := os.ReadFile(from)
				if err != nil {
					return err
				}
				content = string(raw)
			}
			if strings.TrimSpace(content) == "" {
				return fmt.Errorf("snippet body is empty (use --body or --from)")
			}
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()

			var existing Snippet
			found, _ := s.Get(store.BucketSnippets, name, &existing)
			now := time.Now()
			var created time.Time
			if found {
				created = existing.CreatedAt
			} else {
				created = now
			}
			sn := Snippet{
				ID:        name,
				Name:      args[0],
				Body:      content,
				Tags:      splitTags(tags),
				CreatedAt: created,
				UpdatedAt: now,
			}
			if err := s.Put(store.BucketSnippets, name, &sn); err != nil {
				return err
			}
			ui.Success("saved snippet " + name)
			return nil
		},
	}
	c.Flags().StringVar(&body, "body", "", "snippet text")
	c.Flags().StringVar(&from, "from", "", "read snippet body from a file")
	c.Flags().StringVar(&tags, "tags", "", "comma-separated tags")
	return c
}

func snippetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List snippets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()

			var items []Snippet
			_ = s.Iter(store.BucketSnippets, func(_ string, raw []byte) error {
				var sn Snippet
				if err := jsonUnmarshal(raw, &sn); err == nil {
					items = append(items, sn)
				}
				return nil
			})
			sort.Slice(items, func(i, j int) bool {
				return items[i].Name < items[j].Name
			})

			tbl := ui.Table{Headers: []string{"NAME", "TAGS", "UPDATED", "PREVIEW"}}
			for _, sn := range items {
				prev := sn.Body
				if len(prev) > 48 {
					prev = prev[:47] + "…"
				}
				prev = strings.ReplaceAll(prev, "\n", " ")
				tbl.Rows = append(tbl.Rows, []string{
					sn.Name,
					strings.Join(sn.Tags, ", "),
					sn.UpdatedAt.Format("2006-01-02"),
					prev,
				})
			}
			pretty := ui.Headline(fmt.Sprintf("Snippets (%d)", len(items))) + "\n" + tbl.Render()
			return ui.Render(items, pretty)
		},
	}
}

func snippetGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Print a snippet body",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()
			sn, err := findSnippet(s, args[0])
			if err != nil {
				return err
			}
			return ui.Render(sn, sn.Body)
		},
	}
}

func snippetCopyCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "copy <name>",
		Aliases: []string{"cp"},
		Short:   "Copy a snippet body to the clipboard",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()
			sn, err := findSnippet(s, args[0])
			if err != nil {
				return err
			}
			if err := clipboard.WriteAll(sn.Body); err != nil {
				return err
			}
			ui.Success("copied snippet: " + sn.Name)
			return nil
		},
	}
}

func snippetRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <name>",
		Aliases: []string{"delete"},
		Short:   "Delete a snippet",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()
			key := normalizeSnippetName(args[0])
			if !ui.Confirm("Delete snippet '"+args[0]+"'?", false) {
				ui.Warn("aborted")
				return nil
			}
			if err := s.Delete(store.BucketSnippets, key); err != nil {
				return err
			}
			ui.Success("deleted")
			return nil
		},
	}
}

func findSnippet(s *store.Store, name string) (*Snippet, error) {
	key := normalizeSnippetName(name)
	var sn Snippet
	ok, err := s.Get(store.BucketSnippets, key, &sn)
	if err != nil {
		return nil, err
	}
	if ok {
		return &sn, nil
	}
	// fuzzy-ish: first prefix/substring match on keys
	var picked *Snippet
	_ = s.Iter(store.BucketSnippets, func(k string, raw []byte) error {
		if picked != nil {
			return nil
		}
		if strings.Contains(strings.ToLower(k), strings.ToLower(key)) {
			var cand Snippet
			if err := jsonUnmarshal(raw, &cand); err == nil {
				picked = &cand
			}
		}
		return nil
	})
	if picked == nil {
		return nil, fmt.Errorf("snippet %q not found", name)
	}
	return picked, nil
}

func normalizeSnippetName(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func splitTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
