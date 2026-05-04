package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Out is the global stdout writer commands should use.
var Out io.Writer = os.Stdout

// Err is the global stderr writer commands should use.
var Err io.Writer = os.Stderr

// jsonMode toggles the --json global flag.
var jsonMode bool

// SetJSON switches all "Render*" helpers to emit machine-readable JSON.
func SetJSON(v bool) { jsonMode = v }

// JSON reports whether JSON output is currently active.
func JSON() bool { return jsonMode }

// Println prints a styled line to Out (or strips styles automatically when piping).
func Println(s string) { fmt.Fprintln(Out, s) }

// Printf prints a styled formatted line to Out.
func Printf(format string, a ...any) { fmt.Fprintf(Out, format, a...) }

// Render writes either pretty (text) or JSON depending on the global mode.
// `data` is whatever should be marshalled in JSON mode.
// `pretty` is the rendered string for human mode.
func Render(data any, pretty string) error {
	if jsonMode {
		enc := json.NewEncoder(Out)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}
	fmt.Fprintln(Out, pretty)
	return nil
}

// Success / Warn / Danger helpers print a coloured prefix.
func Success(msg string) { fmt.Fprintln(Out, T().Success.Render("✔ ")+msg) }
func Warn(msg string)    { fmt.Fprintln(Out, T().Warning.Render("! ")+msg) }
func Danger(msg string)  { fmt.Fprintln(Err, T().Danger.Render("✘ ")+msg) }
func Info(msg string)    { fmt.Fprintln(Out, T().Info.Render("→ ")+msg) }
