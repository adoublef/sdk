package fs_test

import (
	"encoding/json"
	"strings"
	"testing"

	"go.adoublef.dev/is"
	. "go.adoublef.dev/sdk/fs"
)

func TestParsePath(t *testing.T) {
	tt := map[string]struct {
		s   string
		err error
	}{
		"ParsePath(\"/a/b/c.d\")": {
			s: "/a/b/c.d",
		},
		"ErrPathAbs(\"./a/b/c\")": {
			s:   "./a/b/c",
			err: ErrPathAbs,
		},
		"ErrPathAbs(\"~/a/b/c\")": {
			s:   "~/a/b/c",
			err: ErrPathAbs,
		},
		"ErrPathLen0(\"\")": {
			err: ErrPathLen0,
		},
		"ErrPathAbs(\"../..\")": {
			s:   "../..",
			err: ErrPathAbs,
		},
		"ErrPathAbs(\"..\")": {
			s:   "..",
			err: ErrPathAbs,
		},
		"ParsePath(\"/a\")": {
			s: "/a",
		},
		"ErrPathLen255(\"/1/2/.../255/m\")": {
			s:   strings.Repeat("/n/", 255) + "m",
			err: ErrPathLen255,
		},
		"ErrBaseLen0(\"/a/b/\")": {
			s:   "/a/b/",
			err: ErrBaseLen0,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			is := is.NewRelaxed(t)

			p, err := ParsePath(tc.s)
			is.Err(err, tc.err) // parse path
			t.Log(p)
		})
	}
}

func TestPath_UnmarshalJSON(t *testing.T) {
	tt := map[string]struct {
		s   string
		err error
	}{
		"OK": {
			s: "\"/a.txt\"",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			var (
				is = is.NewRelaxed(t)
			)

			var path Path
			err := json.NewDecoder(strings.NewReader(tc.s)).Decode(&path)
			is.Err(err, tc.err) // decode pathname
		})
	}
}

func TestPath_MarshalJSON(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var (
			is = is.NewRelaxed(t)
		)
		input := "\"/a.txt\""

		var path Path
		err := json.NewDecoder(strings.NewReader(input)).Decode(&path)
		is.NoErr(err) // decode pathname

		var sb strings.Builder
		err = json.NewEncoder(&sb).Encode(path)
		is.NoErr(err) // encode pathname

		is.Equal(strings.TrimSpace(sb.String()), input)
	})
}
