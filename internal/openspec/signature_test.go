package openspec

import (
	"io/fs"
	"os"
	"testing"
	"time"
)

// countingFS is a fileSystem that records how many reads vs stats happen, so we
// can prove Signature only stats (never reads).
type countingFS struct {
	reads, stats int
	exists       bool
	mod          time.Time
	size         int64
}

func (c *countingFS) ReadFile(string) ([]byte, error)             { c.reads++; return nil, nil }
func (c *countingFS) WriteFile(string, []byte, os.FileMode) error { return nil }
func (c *countingFS) ReadDir(string) ([]os.DirEntry, error)       { return nil, nil }
func (c *countingFS) Stat(string) (os.FileInfo, error) {
	c.stats++
	if !c.exists {
		return nil, fs.ErrNotExist
	}
	return fakeInfo{mod: c.mod, size: c.size}, nil
}

type fakeInfo struct {
	mod  time.Time
	size int64
}

func (f fakeInfo) Name() string       { return "f" }
func (f fakeInfo) Size() int64        { return f.size }
func (f fakeInfo) Mode() os.FileMode  { return 0o644 }
func (f fakeInfo) ModTime() time.Time { return f.mod }
func (f fakeInfo) IsDir() bool        { return false }
func (f fakeInfo) Sys() any           { return nil }

func TestSignature(t *testing.T) {
	t.Run("present file: populated from stat, no read", func(t *testing.T) {
		mod := time.Unix(1700000000, 123)
		cfs := &countingFS{exists: true, size: 42, mod: mod}
		sig := NewLoader(cfs).Signature("tasks.md")
		if !sig.Present || sig.Size != 42 || sig.ModUnixNano != mod.UnixNano() {
			t.Errorf("unexpected signature: %+v", sig)
		}
		if cfs.reads != 0 {
			t.Errorf("Signature must not read the file; got %d reads", cfs.reads)
		}
		if cfs.stats != 1 {
			t.Errorf("expected exactly one stat, got %d", cfs.stats)
		}
	})

	t.Run("absent file yields !Present", func(t *testing.T) {
		cfs := &countingFS{exists: false}
		if NewLoader(cfs).Signature("gone.md").Present {
			t.Error("a not-found stat should yield Present=false")
		}
	})

	t.Run("size or mtime change flips the signature", func(t *testing.T) {
		base := NewLoader(&countingFS{exists: true, size: 10, mod: time.Unix(1, 0)}).Signature("x")
		biggerSize := NewLoader(&countingFS{exists: true, size: 11, mod: time.Unix(1, 0)}).Signature("x")
		newerMtime := NewLoader(&countingFS{exists: true, size: 10, mod: time.Unix(2, 0)}).Signature("x")
		if base == biggerSize || base == newerMtime {
			t.Error("changing size or mtime must change the signature")
		}
	})
}
