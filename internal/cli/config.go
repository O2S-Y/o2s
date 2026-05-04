package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/config"
	"github.com/O2S-Y/o2s/internal/ui"
)

func configCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Inspect or edit O2S configuration",
	}
	c.AddCommand(configShowCmd(), configPathCmd(), configResetCmd(), configSetThemeCmd())
	return c
}

func configShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print current configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			pretty := ui.Headline("Configuration") + "\n" + ui.KV([][2]string{
				{"name", orDash(cfg.Name)},
				{"theme", cfg.Theme},
				{"editor", cfg.Editor},
				{"no_color", fmt.Sprintf("%v", cfg.NoColor)},
				{"onboarded", fmt.Sprintf("%v", cfg.Onboard)},
			})
			return ui.Render(cfg, pretty)
		},
	}
}

func configPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the absolute path of the config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := config.Path()
			if err != nil {
				return err
			}
			ui.Println(p)
			return nil
		},
	}
}

func configResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Restore configuration to factory defaults",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !ui.Confirm("Reset O2S configuration to defaults?", false) {
				ui.Warn("aborted")
				return nil
			}
			c := config.Default()
			c.Onboard = true
			if err := config.Save(c); err != nil {
				return err
			}
			ui.Success("config reset")
			return nil
		},
	}
}

func configSetThemeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-theme [name]",
		Short: "Persist a colour theme as default",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			var picked string
			if len(args) == 1 {
				picked = args[0]
			} else {
				picked, err = ui.Select("Pick a theme", ui.Themes())
				if err != nil {
					return err
				}
			}
			cfg.Theme = picked
			if err := config.Save(cfg); err != nil {
				return err
			}
			ui.SetTheme(picked)
			ui.Success("theme set to " + picked)
			return nil
		},
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
