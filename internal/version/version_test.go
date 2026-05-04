package version

import (
	"strings"
	"testing"
)

func TestFullDefaults(t *testing.T) {
	got := Full()
	for _, want := range []string{"dev", "none", "unknown"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Full() = %q, want substring %q", got, want)
		}
	}
}
