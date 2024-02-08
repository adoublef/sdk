package strconv_test

import (
	"testing"

	. "go.adoublef.dev/sdk/strconv"
)

func TestFormatIEC(t *testing.T) {
	tt := map[string]struct {
		n    uint
		want string
	}{
		"0B": {
			n:    0,
			want: "0B",
		},
		"27B": {
			n:    27,
			want: "27B",
		},
		"999B": {
			n:    999,
			want: "999B",
		},
		"1000B": {
			n:    1000,
			want: "1000B",
		},
		"1023B": {
			n:    1023,
			want: "1023B",
		},
		"1024B": {
			n:    1024,
			want: "1.0KiB",
		},
		"1728B": {
			n:    1728,
			want: "1.7KiB",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if iec := FormatIEC(tc.n); tc.want != iec {
				t.Errorf("expected %q; got %q", tc.want, iec)
			}
		})
	}
}
