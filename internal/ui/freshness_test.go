package ui

import (
	"testing"

	"github.com/danshort/lectern/internal/openspec"
)

func TestFreshnessChanged(t *testing.T) {
	f := newFreshness()
	a := openspec.FileSig{Present: true, ModUnixNano: 1, Size: 10}

	if !f.changed("p", a) {
		t.Error("first sight of a path should count as changed")
	}
	if f.changed("p", a) {
		t.Error("identical signature should be unchanged")
	}

	bMtime := openspec.FileSig{Present: true, ModUnixNano: 2, Size: 10}
	if !f.changed("p", bMtime) {
		t.Error("a different mtime should count as changed")
	}

	bSize := openspec.FileSig{Present: true, ModUnixNano: 2, Size: 11}
	if !f.changed("p", bSize) {
		t.Error("a different size should count as changed")
	}

	if !f.changed("p", openspec.FileSig{}) {
		t.Error("present → absent (deletion) should count as changed")
	}
	if f.changed("p", openspec.FileSig{}) {
		t.Error("still-absent should be unchanged")
	}

	// Distinct paths are tracked independently.
	if !f.changed("q", a) {
		t.Error("a new path should count as changed regardless of other paths")
	}
}
