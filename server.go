package main // templates

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Slave struct {
	files      []string
	connection net.Conn
}

var slaves []Slave

// func sendPasswordToSlaves(password) {
//
// }

func handleSlaveConnection(c net.Conn, addchan chan<- Slave, rmchan chan<- Slave) {
	buf := make([]byte, 4096)
	defer c.Close()
	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		files := string(buf[0:n])

		filesArray := strings.Split(files, ",")
		currentSlave := Slave{filesArray, c}
		slaves = append(slaves, currentSlave)
		fmt.Println(filesArray)
		addchan <- currentSlave
		fmt.Println(filesArray)
		defer func() {
			fmt.Println("Slave has left.")
			log.Printf("Connection from %v closed.\n", c.RemoteAddr())
			rmchan <- currentSlave
		}()

	}
}

func handleSlaves() {

	addchan := make(chan Slave)
	rmchan := make(chan Slave)

	ln, err := net.Listen("tcp", "127.0.0.1:8002")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleSlaveConnection(conn, addchan, rmchan)
	}
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
		// sendPasswordToSlaves(password)
	}
}

func main() {

	go handleSlaves()

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
