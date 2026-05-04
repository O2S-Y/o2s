package files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/O2S-Y/o2s/internal/ui"
)

// convertCmd handles json <-> yaml conversion. We intentionally keep the
// scope tight (the two by far most common cases) and document where to
// extend later without bringing in heavyweight deps.
func convertCmd() *cobra.Command {
	var (
		toFmt string
		out   string
	)
	c := &cobra.Command{
		Use:   "convert <file>",
		Short: "Convert between JSON and YAML",
		Long:  "Convert between JSON and YAML.  Source format is detected from extension, target is set with --to.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(src), "."))
			from := ext
			if from == "yml" {
				from = "yaml"
			}
			to := strings.ToLower(toFmt)
			if to == "" {
				switch from {
				case "json":
					to = "yaml"
				case "yaml":
					to = "json"
				default:
					return fmt.Errorf("cannot infer target; use --to {json|yaml}")
				}
			}

			raw, err := os.ReadFile(src)
			if err != nil {
				return err
			}
			var data any
			switch from {
			case "json":
				if err := json.Unmarshal(raw, &data); err != nil {
					return fmt.Errorf("decode json: %w", err)
				}
			case "yaml":
				if err := yaml.Unmarshal(raw, &data); err != nil {
					return fmt.Errorf("decode yaml: %w", err)
				}
			default:
				return fmt.Errorf("unsupported source format: %s", from)
			}

			var encoded []byte
			switch to {
			case "json":
				encoded, err = json.MarshalIndent(data, "", "  ")
			case "yaml":
				encoded, err = yaml.Marshal(data)
			default:
				return fmt.Errorf("unsupported target format: %s", to)
			}
			if err != nil {
				return err
			}

			if out == "" {
				ui.Println(string(encoded))
				return nil
			}
			if err := os.WriteFile(out, encoded, 0o644); err != nil {
				return err
			}
			ui.Success("wrote " + out)
			return nil
		},
	}
	c.Flags().StringVar(&toFmt, "to", "", "target format: json | yaml")
	c.Flags().StringVarP(&out, "out", "o", "", "output path (default: stdout)")
	return c
}
