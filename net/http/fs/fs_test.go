package fs_test

import (
	"embed"
	"testing"

	"github.com/matryer/is"
	"go.adoublef.dev/sdk/http/fs"
)

//go:embed all:testdata
var embedFS embed.FS

// var fsys = fs.Must(embedFS, "testdata")

func TestFS(t *testing.T) {
	t.Run("sub-root of file system", func(t *testing.T) {
		is := is.New(t)

		fs, err := fs.Sub(embedFS, "testdata")
		is.NoErr(err) // subtree of embedded file system

		_, err = fs.Open("a")
		is.NoErr(err) // open file `a.html` using the filename
	})
}
