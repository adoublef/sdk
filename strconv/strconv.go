package strconv

import (
	"fmt"
	"strconv"
)

// Quote returns a double-quoted Go string literal representing s.
func Quote(s fmt.Stringer) string {
	return strconv.Quote(s.String())
}

// Utoa is equivalent to FormatUInt(uint64(u), 10).
func Utoa(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}

// Atoi is equivalent to ParseUint(s, 10, 0), converted to type uint.
func Atou(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 0)
	return uint(n), err
}

// FormatIEC returns the string representation of u using the IEC standard.
func FormatIEC(u uint) string {
	const unit = 1024
	if u < unit {
		return fmt.Sprintf("%dB", u)
	}
	div, exp := int64(unit), 0
	for n := u / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	// NOTE look to using strings.Builder instead
	return fmt.Sprintf("%.1f%ciB", float64(u)/float64(div), "KMGTPE"[exp])
}

// FormatIEC returns the string representation of u using the SI standard.
func FormatSI(u uint) string {
	const unit = 1000
	if u < unit {
		return fmt.Sprintf("%dB", u)
	}
	div, exp := int64(unit), 0
	for n := u / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	// NOTE look to using strings.Builder instead
	return fmt.Sprintf("%.1f%cB", float64(u)/float64(div), "kMGTPE"[exp])
}
