package main // templates

import (
	"fmt"
	"net"
)

func sendPasswordToSlaves(password) {

}

func handleConnection(c net.Conn) {
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		password := string(buf[0:n])

		fmt.Println(password)
		sendPasswordToSlaves(password)
	}
}

func main() {

	ln, err := net.Listen("tcp", "127.0.0.1:8003")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}

}
