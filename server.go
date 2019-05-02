package main // templates

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Slave struct {
	files []string
	conn  net.Conn
}

var slaves []Slave

func stopSearchWhenFound(c net.Conn, finishSearchChan chan bool) {
	for {
		select {
		case searchFinished := <-finishSearchChan:
			if searchFinished {
				c.Write([]byte("You Password has been PWNED!!!"))
				c.Close()
			}
		}
	}
}

func sendPasswordToSlaves(password string, finishSearchChan chan bool) {
	buf := make([]byte, 4096)
	for _, s := range slaves {
		s.conn.Write([]byte("s:" + password + ":" + s.files[0]))
		n, err := s.conn.Read(buf)
		if err != nil || n == 0 {
			s.conn.Close()
			break
		}
		result := string(buf[0:n])

		if result == "1" {
			finishSearchChan <- true
		}

	}
}

func handleSlaveConnection(c net.Conn, addchan chan Slave, rmchan chan Slave) {
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

		// addchan <- currentSlave
		// fmt.Println(filesArray)
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

func handleClientConnection(c net.Conn) {
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		password := string(buf[0:n])

		fmt.Println(password)

		finishSearchChan := make(chan bool)
		go sendPasswordToSlaves(password, finishSearchChan)
		go stopSearchWhenFound(c, finishSearchChan)

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
		go handleClientConnection(conn)
	}

}
