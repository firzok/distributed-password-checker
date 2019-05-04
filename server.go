package main // templates

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Slave struct {
	files         []string
	conn          net.Conn
	passwordFound bool
	freeToSearch  bool
}

type Client struct {
	password string
	conn     net.Conn
}

var clients = make(map[net.Conn]Client)
var slaves = make(map[net.Conn]Slave)

func stopSearchWhenFound(c net.Conn, finishSearchChan chan bool) {
	for {
		select {
		case searchFinished := <-finishSearchChan:
			// fmt.Println("Search Finished")
			if searchFinished {
				// fmt.Println("Password found")
				c.Write([]byte("1"))
				c.Close()
			}
		}
	}
}

func sendPasswordToSlaves(password string) {
	for k, v := range slaves {
		k.Write([]byte("s:" + password + ":" + v.files[0]))
	}
}

func passwordFound(password string) {
	for _, c := range clients {
		if c.password == password {
			fmt.Println("Result sent to Client")
			c.conn.Write([]byte("pf"))
		}
	}
}

func passwordNotFound(password string) {
	for _, c := range clients {
		if c.password == password {
			fmt.Println("Result sent to Client")
			c.conn.Write([]byte("pnf"))
		}
	}
}

func handleSlaveConnection(c net.Conn) {

	//Get slave filesArray
	buf := make([]byte, 4096)
	defer c.Close()
	n, err := c.Read(buf)
	if err != nil || n == 0 {
		c.Close()
	}
	files := string(buf[0:n])

	filesArray := strings.Split(files, ",")
	currentSlave := Slave{filesArray, c, false, true}
	slaves[c] = currentSlave
	fmt.Println("Files: ", filesArray)

	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		message := string(buf[0:n])
		fmt.Println("Message from Slave: ", message)

		msgSplits := strings.Split(message, ":")

		//pf = Password Found
		if msgSplits[0] == "pf" {
			fmt.Println("Password Found")
			passwordFound(msgSplits[1])
		} else if msgSplits[0] == "pnf" {
			fmt.Println("Password NOT Found")
			passwordNotFound(msgSplits[1])
		}

		defer func() {
			fmt.Println("Slave has left.")
			log.Printf("Connection from %v closed.\n", c.RemoteAddr())
			delete(slaves, c)
		}()

	}
}

func handleSlaves() {

	ln, err := net.Listen("tcp", "127.0.0.1:8002")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleSlaveConnection(conn)
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

		fmt.Println("Client Sent: ", password)

		currentClient := Client{password, c}
		//add to map
		clients[c] = currentClient

		go sendPasswordToSlaves(password)

		defer func() {
			fmt.Println("Client has left.")
			log.Printf("Connection from %v closed.\n", c.RemoteAddr())
			delete(clients, c)
		}()
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
