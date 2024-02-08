package iotest_test

import (
	"io"
	"testing"

	"go.adoublef.dev/is"
	"go.adoublef.dev/sdk/io/fs"
	. "go.adoublef.dev/sdk/io/iotest"
)

func TestReader_Read(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		r    *Reader
		n    int
	}{
		{
			name: "1024 Bytes",
			r:    NewReader(fs.KB.Int()),
			n:    fs.KB.Int(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			is := is.NewRelaxed(t)

			n, err := io.ReadAll(tc.r)
			is.NoErr(err) // read all
			is.Equal(len(n), tc.n)
		})
	}
}
