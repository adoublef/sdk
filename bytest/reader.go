// source https://medium.com/@nicksoetaert/using-io-reader-to-simulate-an-arbitrary-sized-file-in-golang-4df6287cae51
package bytest

import (
	"io"
)

const (
	B = 1 << (10 * iota)
	KB
	MB
	GB // 1048576000 1073741824
)

func NewReader(n int) *Reader {
	return &Reader{n: n}
}

// Reader simulates a fake read-only file of an arbitrary size.
// The contents of this file cycles through the lowercase alphabetical ascii characters (abcdefghijklmnopqrstuvwxyz)
// Data is generated as requested, so it is safe to generate a 100GB file with 1GB of memory free on your machine.
type Reader struct {
	n   int
	off int
}

// Read reads lowercase ascii characters into b from f.
// n is number of bytes read.
// If a Read is attempted at the end of the file, io.EOF is returned.
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.n == r.off {
		return 0, io.EOF
	}
	m := len(p)
	if (r.n - r.off) < m {
		m = r.n - r.off
	}
	for i := 0; i < m; i++ {
		p[i] = byte('a' + (r.off+i)%26)
	}
	r.off += m
	return m, nil
}
