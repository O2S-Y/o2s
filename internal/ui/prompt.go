package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
)

// Confirm prompts the user for yes/no using a styled Huh form.
// Falls back to a stdin read for non-TTY environments (CI, pipes).
func Confirm(question string, def bool) bool {
	if !IsTTY() {
		fmt.Fprintf(os.Stderr, "%s [y/N]: ", question)
		r := bufio.NewReader(os.Stdin)
		line, _ := r.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			return def
		}
		return line == "y" || line == "yes"
	}
	out := def
	_ = huh.NewConfirm().
		Title(question).
		Affirmative("Yes").
		Negative("No").
		Value(&out).
		Run()
	return out
}

// Select prompts the user to pick one option from the list.
func Select(title string, options []string) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options")
	}
	if !IsTTY() {
		return options[0], nil
	}
	var picked string
	opts := make([]huh.Option[string], 0, len(options))
	for _, o := range options {
		opts = append(opts, huh.NewOption(o, o))
	}
	err := huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Value(&picked).
		Run()
	return picked, err
}

// Input asks for a single line of text input.
func Input(label, placeholder, def string) (string, error) {
	if !IsTTY() {
		return def, nil
	}
	v := def
	err := huh.NewInput().
		Title(label).
		Placeholder(placeholder).
		Value(&v).
		Run()
	return v, err
}
