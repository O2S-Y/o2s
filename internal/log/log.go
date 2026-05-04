// Package log centralises the application logger.
package log

import (
	"os"

	clog "github.com/charmbracelet/log"
)

var L = clog.NewWithOptions(os.Stderr, clog.Options{
	ReportTimestamp: false,
	Prefix:          "o2s",
})

func SetLevel(verbose bool) {
	if verbose {
		L.SetLevel(clog.DebugLevel)
		return
	}
	L.SetLevel(clog.InfoLevel)
}
