package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"

	"github.com/O2S-Y/o2s/internal/config"
	"github.com/O2S-Y/o2s/internal/ui"
)

// runOnboarding shows a one-time Huh form on first launch to capture
// the user's preferred theme, editor, and display name.
func runOnboarding(cfg *config.Config) error {
	if !ui.IsTTY() {
		cfg.Onboard = true
		return nil
	}
	fmt.Fprintln(os.Stdout, ui.Banner())
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, ui.Headline("Welcome — let's set things up"))

	themeOpts := make([]huh.Option[string], 0, len(ui.Themes()))
	for _, t := range ui.Themes() {
		themeOpts = append(themeOpts, huh.NewOption(t, t))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What should I call you?").
				Placeholder("e.g. Sami").
				Value(&cfg.Name),
			huh.NewSelect[string]().
				Title("Pick a theme").
				Options(themeOpts...).
				Value(&cfg.Theme),
			huh.NewInput().
				Title("Default editor").
				Placeholder("e.g. code, nvim, notepad").
				Value(&cfg.Editor),
			huh.NewConfirm().
				Title("Disable colour output?").
				Affirmative("Yes (NO_COLOR)").
				Negative("No, keep colours").
				Value(&cfg.NoColor),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}
	cfg.Onboard = true
	ui.Success("you're all set, " + firstNonEmpty(cfg.Name, "friend"))
	return nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
