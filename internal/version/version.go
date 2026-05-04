// Package version exposes build-time identifiers (set via -ldflags).
package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func Full() string {
	return Version + " (" + Commit + ", " + Date + ")"
}
