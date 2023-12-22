package bytest_test

import (
	"io"
	"testing"

	"github.com/adoublef/sdk/bytest"
	"github.com/matryer/is"
)

func TestReader(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		r    *bytest.Reader
		n    int
	}{
		{
			name: "1024 Bytes",
			r:    bytest.NewReader(bytest.KB),
			n:    bytest.KB,
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
