package ids

import (
	"strings"
	"testing"
)

func TestNewIsUnique(t *testing.T) {
	a, b := New(), New()
	if a == b {
		t.Fatalf("expected unique IDs, got %s twice", a)
	}
	if len(a) != 26 {
		t.Fatalf("expected 26-char ULID, got %d (%q)", len(a), a)
	}
}

func TestNewIsLexicographicallySortable(t *testing.T) {
	first := New()
	// Sleep a tiny bit so the millisecond-precision timestamp ticks over.
	for i := 0; i < 10; i++ {
		_ = New()
	}
	last := New()
	if strings.Compare(first, last) >= 0 {
		t.Fatalf("expected first (%s) < last (%s)", first, last)
	}
}
