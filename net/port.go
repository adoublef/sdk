package net

import (
	"fmt"
	"net"
)

// Next will return the next freely available tcp port
func Next() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("net: failed to listen on a port: %v", err))
		}
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}
