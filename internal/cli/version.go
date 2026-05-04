package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
	"github.com/O2S-Y/o2s/internal/version"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show O2S version & build info",
		RunE: func(cmd *cobra.Command, _ []string) error {
			data := map[string]string{
				"version": version.Version,
				"commit":  version.Commit,
				"date":    version.Date,
				"go":      runtime.Version(),
				"os":      runtime.GOOS,
				"arch":    runtime.GOARCH,
			}
			pretty := ui.Banner() + "\n\n" + ui.KV([][2]string{
				{"version", version.Version},
				{"commit", version.Commit},
				{"built", version.Date},
				{"go", runtime.Version()},
				{"os/arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)},
			})
			return ui.Render(data, pretty)
		},
	}
}
