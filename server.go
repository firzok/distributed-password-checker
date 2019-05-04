package main // templates

import (
	"fmt"
	"log"
	"net"
	"strings"
)

//Slave is struct for currently connected slaves
type Slave struct {
	files        []string //Files that slave have
	conn         net.Conn //connection for communication
	freeToSearch bool     //either free to search or currently searching
}

//Client is struct for clients being currently connected to server
type Client struct {
	password string   //Password to be searched
	conn     net.Conn // Connection for sending result back to client
}

var clients = make(map[net.Conn]Client)
var slaves = make(map[net.Conn]Slave)

//all files that all slaves have and value is that either they have been searched or not
var slaveFiles = make(map[string]bool)

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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func sendPasswordToSlaves(password string) {

	//loop over all slaves
	for k, v := range slaves {

		//if that slave is free to search
		if v.freeToSearch {

			//loop over all the Slave files
			for fileName, searched := range slaveFiles {

				//if that file has not yet been searched and the slave has that file then search that file for password
				if !searched && stringInSlice(fileName, v.files) {

					fmt.Println("Password sent to slave", fileName)
					//sending command to search for password in that file
					k.Write([]byte("s:" + password + ":" + fileName))

					//not free to search anymore
					t := slaves[k]
					t.freeToSearch = false
					slaves[k] = t

					//file has been searched
					slaveFiles[fileName] = true
					break
				}
			}
		}
	}
}

func passwordFoundBySlave(password string, slaveConn net.Conn) {
	for _, c := range clients {
		if c.password == password {
			fmt.Println("Result sent to Client")
			c.conn.Write([]byte("pf"))
		}
	}

	//Telling all slaves(except the one that found it) that password is found so stop searching
	for k := range slaves {
		//all slaves back to free to search for new search
		t := slaves[k]
		t.freeToSearch = true
		slaves[k] = t
		if k != slaveConn {
			fmt.Println("Stop search request sent to slave")
			k.Write([]byte("pf:" + password))
		}
	}

	//all files back to not searched for new search
	for k := range slaveFiles {
		slaveFiles[k] = false
	}

}

func passwordNotFound(password string) {

	//loop over all slave files
	for _, v := range slaveFiles {

		//if any file has not yet been searched
		if !v {
			//send password to be searched
			sendPasswordToSlaves(password)
			return
		}
	}

	//all files have been searched and still password not Found
	//loop over all clients
	for _, c := range clients {

		//if client asked ofr this password
		if c.password == password {

			fmt.Println("Result sent to Client")
			//tell client that password not found
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
	currentSlave := Slave{filesArray, c, true}
	slaves[c] = currentSlave
	fmt.Println("Files: ", filesArray)
	for _, f := range filesArray {
		slaveFiles[f] = false
	}

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
			passwordFoundBySlave(msgSplits[1], c)
		} else if msgSplits[0] == "pnf" {
			fmt.Println("Password NOT Found")

			//setting slave free again to search
			for k := range slaves {
				if slaves[k].conn == c {
					t := slaves[k]
					t.freeToSearch = true
					slaves[k] = t
				}
			}

			passwordNotFound(msgSplits[1])

		}

		defer func() {
			log.Printf("Connection from Slave %v closed.\n", c.RemoteAddr())
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
			log.Printf("Connection from Client %v closed.\n", c.RemoteAddr())
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
