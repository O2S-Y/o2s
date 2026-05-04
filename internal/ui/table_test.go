package ui

import (
	"strings"
	"testing"
)

func TestTableRendersWithHeadersAndRows(t *testing.T) {
	out := Table{
		Headers: []string{"NAME", "AGE"},
		Rows:    [][]string{{"alice", "30"}, {"bob", "25"}},
	}.Render()

	for _, want := range []string{"NAME", "AGE", "alice", "bob"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestTableEmpty(t *testing.T) {
	if got := (Table{}).Render(); got != "" {
		t.Fatalf("expected empty string for empty table, got %q", got)
	}
}
