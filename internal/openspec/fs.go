package openspec

import "os"

type fileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	ReadDir(name string) ([]os.DirEntry, error)
	Stat(name string) (os.FileInfo, error)
}

type OSFS struct{}

// dossier is a local, single-user tool that reads and writes files in the
// user's own project at the user's own privilege level. Paths are built from
// fixed artifact names joined to directory entries discovered via ReadDir, so
// there is no privilege boundary or untrusted input being crossed here.

// #nosec G304 -- see package note above: reads the user's own project files.
func (OSFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

// #nosec G304 -- see package note above: writes the user's own project files.
func (OSFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (OSFS) ReadDir(name string) ([]os.DirEntry, error) { return os.ReadDir(name) }
func (OSFS) Stat(name string) (os.FileInfo, error)      { return os.Stat(name) }
