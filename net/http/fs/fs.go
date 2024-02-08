package fs

import (
	"fmt"
	"io/fs"
	"os"
)

var _ fs.FS = (FS{})

// An FS provides access to a hierarchical file system.
type FS struct {
	fsys fs.FS
}

// Must panics if there is an error returning the subtree
func Must(fsys fs.FS, name string) FS {
	f, err := Sub(fsys, name)
	if err != nil {
		panic(err)
	}
	return f
}

// Sub returns an FS corresponding to the subtree rooted at fsys's dir.
func Sub(fsys fs.FS, name string) (FS, error) {
	f, err := fs.Sub(fsys, name)
	if err != nil {
		return FS{}, fmt.Errorf("sub: %w", err)
	}
	return FS{fsys: f}, nil
}

// Open tries to open the file with the supplied name. If not found retires with `.html` extension
func (fsys FS) Open(name string) (fs.File, error) {
	f, err := fsys.fsys.Open(name)
	if os.IsNotExist(err) {
		if f, err := fsys.fsys.Open(name + ".html"); err == nil {
			return f, nil
		}
	}
	return f, err
}

func NewFS(fsys fs.FS) *FS {
	return &FS{fsys: fsys}
}
