package dev

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func envCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "env",
		Short: "Work with dotenv (.env) files",
	}
	c.AddCommand(envPrintCmd(), envGetCmd())
	return c
}

func envPrintCmd() *cobra.Command {
	var (
		format string
		path   string
		export bool
	)
	c := &cobra.Command{
		Use:     "print [path]",
		Aliases: []string{"dump", "show"},
		Short:   "Load and print variables from a .env file",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := path
			if target == "" && len(args) == 1 {
				target = args[0]
			}
			if target == "" {
				target = ".env"
			}
			m, err := readEnvFile(target)
			if err != nil {
				return err
			}
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			if export && format == "posix" {
				for _, k := range keys {
					fmt.Fprintf(ui.Out, "export %s=%s\n", k, posixQuote(m[k]))
				}
				return nil
			}

			switch format {
			case "json":
				return ui.Render(m, "")
			case "cmd":
				for _, k := range keys {
					fmt.Fprintf(ui.Out, "set %s=%s\n", k, cmdEscape(m[k]))
				}
				return nil
			case "powershell":
				for _, k := range keys {
					fmt.Fprintf(ui.Out, "$env:%s=\"%s\"\n", k, psEscape(m[k]))
				}
				return nil
			case "posix":
				for _, k := range keys {
					fmt.Fprintf(ui.Out, "%s=%s\n", k, posixQuote(m[k]))
				}
				return nil
			case "table":
				tbl := ui.Table{Headers: []string{"KEY", "VALUE"}}
				for _, k := range keys {
					tbl.Rows = append(tbl.Rows, []string{k, m[k]})
				}
				ui.Println(ui.Headline("Env — "+target) + "\n" + tbl.Render())
				return nil
			default:
				return fmt.Errorf("unknown --format %q", format)
			}
		},
	}
	c.Flags().StringVarP(&path, "file", "f", "", ".env path (default: .env or first positional arg)")
	c.Flags().StringVar(&format, "format", "table", "table | json | posix | cmd | powershell")
	c.Flags().BoolVar(&export, "export", false, "with --format posix, prefix lines with export")
	return c
}

func envGetCmd() *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "get <key> [path]",
		Short: "Print a single variable value",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			target := path
			if target == "" && len(args) == 2 {
				target = args[1]
			}
			if target == "" {
				target = ".env"
			}
			m, err := readEnvFile(target)
			if err != nil {
				return err
			}
			v, ok := m[key]
			if !ok {
				return fmt.Errorf("%q not found in %s", key, target)
			}
			return ui.Render(map[string]string{key: v}, v)
		},
	}
	c.Flags().StringVarP(&path, "file", "f", "", ".env path (default: .env)")
	return c
}

func readEnvFile(path string) (map[string]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return godotenv.Read(path)
}

func posixQuote(v string) string {
	if v == "" {
		return "''"
	}
	if strings.ContainsAny(v, " \t\n\"'$`\\") {
		return fmt.Sprintf("%q", v)
	}
	return v
}

func cmdEscape(v string) string {
	return strings.ReplaceAll(v, "^", "^^")
}

func psEscape(v string) string {
	return strings.ReplaceAll(strings.ReplaceAll(v, "`", "``"), "\"", "`\"")
}
