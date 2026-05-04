package prod

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/store"
	"github.com/O2S-Y/o2s/internal/ui"
)

type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	DoneAt    time.Time `json:"done_at,omitempty"`
	Priority  string    `json:"priority,omitempty"` // low / med / high
}

func todoCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "todo",
		Aliases: []string{"t"},
		Short:   "Lightweight todo list",
	}
	c.AddCommand(todoAddCmd(), todoListCmd(), todoDoneCmd(), todoRmCmd())
	return c
}

func todoAddCmd() *cobra.Command {
	var prio string
	c := &cobra.Command{
		Use:   "add <title...>",
		Short: "Add a new todo",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()
			title := strings.Join(args, " ")
			t := Todo{
				ID:        time.Now().UTC().Format("20060102T150405.000000"),
				Title:     title,
				CreatedAt: time.Now(),
				Priority:  prio,
			}
			if err := s.Put(store.BucketTodos, t.ID, &t); err != nil {
				return err
			}
			ui.Success("added: " + title)
			return nil
		},
	}
	c.Flags().StringVarP(&prio, "priority", "p", "med", "low | med | high")
	return c
}

func todoListCmd() *cobra.Command {
	var all bool
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List todos",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()
			todos := loadAllTodos(s)
			if !all {
				active := todos[:0]
				for _, t := range todos {
					if !t.Done {
						active = append(active, t)
					}
				}
				todos = active
			}
			sort.Slice(todos, func(i, j int) bool {
				return todos[i].CreatedAt.After(todos[j].CreatedAt)
			})

			th := ui.T()
			if len(todos) == 0 {
				ui.Println(th.Muted.Render("nothing here — add one with `o2s prod todo add ...`"))
				return nil
			}
			tbl := ui.Table{
				Headers: []string{"", "ID", "PRIO", "TITLE", "AGE"},
				Aligns:  []string{"left", "left", "left", "left", "right"},
			}
			for _, t := range todos {
				mark := th.Warning.Render("○")
				if t.Done {
					mark = th.Success.Render("●")
				}
				prio := prioBadge(t.Priority)
				age := time.Since(t.CreatedAt).Round(time.Minute).String()
				tbl.Rows = append(tbl.Rows, []string{
					mark, shortID(t.ID), prio, t.Title, age,
				})
			}
			pretty := ui.Headline(fmt.Sprintf("Todos (%d)", len(todos))) + "\n" + tbl.Render()
			return ui.Render(todos, pretty)
		},
	}
	c.Flags().BoolVarP(&all, "all", "a", false, "include completed todos")
	return c
}

func todoDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "done <id>",
		Aliases: []string{"complete"},
		Short:   "Mark a todo as done",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()
			t, err := findTodo(s, args[0])
			if err != nil {
				return err
			}
			t.Done = true
			t.DoneAt = time.Now()
			if err := s.Put(store.BucketTodos, t.ID, t); err != nil {
				return err
			}
			ui.Success("done: " + t.Title)
			return nil
		},
	}
}

func todoRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <id>",
		Aliases: []string{"delete"},
		Short:   "Delete a todo",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()
			t, err := findTodo(s, args[0])
			if err != nil {
				return err
			}
			if !ui.Confirm("Delete '"+t.Title+"' ?", false) {
				ui.Warn("aborted")
				return nil
			}
			if err := s.Delete(store.BucketTodos, t.ID); err != nil {
				return err
			}
			ui.Success("deleted")
			return nil
		},
	}
}

// loadAllTodos returns every persisted todo (used by list + dashboard).
func loadAllTodos(s *store.Store) []Todo {
	var out []Todo
	_ = s.Iter(store.BucketTodos, func(k string, raw []byte) error {
		var t Todo
		if err := jsonUnmarshal(raw, &t); err == nil {
			out = append(out, t)
		}
		return nil
	})
	return out
}

func findTodo(s *store.Store, idOrPrefix string) (*Todo, error) {
	var match *Todo
	_ = s.Iter(store.BucketTodos, func(k string, raw []byte) error {
		if !strings.HasPrefix(k, idOrPrefix) && !strings.HasPrefix(shortID(k), idOrPrefix) {
			return nil
		}
		var t Todo
		if err := jsonUnmarshal(raw, &t); err != nil {
			return nil
		}
		match = &t
		return nil
	})
	if match == nil {
		return nil, fmt.Errorf("no todo matched %q", idOrPrefix)
	}
	return match, nil
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[len(id)-8:]
}

func prioBadge(p string) string {
	t := ui.T()
	switch strings.ToLower(p) {
	case "high", "h":
		return t.Danger.Render("HIGH")
	case "low", "l":
		return t.Muted.Render("low ")
	default:
		return t.Warning.Render("med ")
	}
}
