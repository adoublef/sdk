package fs

import "go.adoublef.dev/sdk/strconv"

// Size
type Size uint

const (
	B Size = 1 << (10 * iota)
	KB
	MB
	GB
)

// Int reutrns Size as an int primitive value.
func (s Size) Int() int {
	return int(s)
}

func (s Size) String() string {
	return strconv.Utoa(uint(s))
}

// IEC returns Size as a formated string using the IEC standard
func (s Size) IEC() string {
	return strconv.FormatIEC(uint(s))
}
