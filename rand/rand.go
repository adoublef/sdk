package rand

import (
	"math/rand"
	"time"
	"unsafe"
)

var src = rand.NewSource(time.Now().UnixNano())

const (
	set  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bits = 6           // 6 bits to represent a letter index
	mask = 1<<bits - 1 // All 1-bits, as many as letterIdxBits
	max  = 63 / bits   // # of letter indices fitting in 63 bits
)

// Bytes returns a random byte slice of length n
func Bytes(n int) []byte {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), max; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), max
		}
		if idx := int(cache & mask); idx < len(set) {
			b[i] = set[idx]
			i--
		}
		cache >>= bits
		remain--
	}

	return b
}

// String returns a random string of length n
func String(n int) string {
	b := Bytes(n)
	return *(*string)(unsafe.Pointer(&b)) // equiv. string(b)
}

// TODO var Reader
