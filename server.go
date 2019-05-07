package main // templates

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

//Slave is struct for currently connected slaves
type Slave struct {
	files             []string //Files that slave have
	conn              net.Conn //connection for communication
	freeToSearch      bool     //either free to search or currently searching
	alive             bool     //set on the basis on periodic heartbeats
	currentSearchFile string   //name of the current file slave is searching
}

//Client is struct for clients being currently connected to server
type Client struct {
	password    string          //Password to be searched
	conn        net.Conn        // Connection for sending result back to client
	searchFiles map[string]bool //files that are yet to be searched for this client
}

var clients []Client
var slaves []Slave

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

func sendPasswordToSlaves(client Client) {

	password := client.password
	fmt.Println("test", password)

	//loop over all slaves
	for slaveIndex, slave := range slaves {

		//if that slave is free to search
		if slave.freeToSearch {
			fmt.Println("test", "free to search")

			//loop over all the Slave files
			for fileName, searched := range client.searchFiles {
				fmt.Println("test", searched, stringInSlice(fileName, slave.files))

				//if that file has not yet been searched and the slave has that file then search that file for password
				if !searched && stringInSlice(fileName, slave.files) {
					//sending command to search for password in that file
					slave.conn.Write([]byte("s:" + password + ":" + fileName))
					fmt.Println("Password sent to slave to be searched in", fileName)

					//not free to search anymore

					slaves[slaveIndex].freeToSearch = false
					slaves[slaveIndex].currentSearchFile = fileName

					//file has been searched
					client.searchFiles[fileName] = true
					break
				}
			}
		}
	}
}

func passwordFoundBySlave(password string, slaveConn net.Conn) {
	for _, client := range clients {
		if client.password == password {
			fmt.Println("Result sent to Client")
			client.conn.Write([]byte("pf"))
		}
	}

	//Telling all slaves(except the one that found it) that password is found so stop searching
	for slaveIndex, slave := range slaves {
		//all slaves back to free to search for new search
		slaves[slaveIndex].freeToSearch = true
		slaves[slaveIndex].currentSearchFile = ""
		if slave.conn != slaveConn {
			fmt.Println("Stop search request sent to slave")
			slave.conn.Write([]byte("pf:" + password))
		}
	}
	//all files back to not searched for new search
	// for k := range slaveFiles {
	// 	slaveFiles[k] = false
	// }

}

func passwordNotFound(password string) {

	var currentClientIndex = 0
	for i, client := range clients {
		if client.password == password {
			currentClientIndex = i
		}
	}

	//loop over all search files
	for _, file := range clients[currentClientIndex].searchFiles {

		//if any file has not yet been searched
		if !file {
			//send password to be searched
			sendPasswordToSlaves(clients[currentClientIndex])
			return
		}
	}

	//all files have been searched and still password not Found
	//loop over all clients
	for _, c := range clients {

		//if client asked for this password
		if c.password == password {

			fmt.Println("Result sent to Client")
			//tell client that password not found
			c.conn.Write([]byte("pnf"))
		}
	}

	for k := range slaves {
		//all slaves back to free to search for new search
		slaves[k].freeToSearch = true
		slaves[k].currentSearchFile = ""

	}
	//all files back to not searched for new search
	for k := range slaveFiles {
		slaveFiles[k] = false
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
	currentSlave := Slave{filesArray, c, true, true, ""}

	slaves = append(slaves, currentSlave)

	fmt.Println("Files from Slaves: ", filesArray)
	for _, f := range filesArray {
		if _, ok := slaveFiles[f]; !ok {
			slaveFiles[f] = false
		}

	}

	// fmt.Println("slaveFiles", slaveFiles)

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
			// slaveFiles[msgSplits[2]] = true
			passwordFoundBySlave(msgSplits[1], c)
		} else if msgSplits[0] == "pnf" { //password not found
			fmt.Println("Password NOT Found")
			// slaveFiles[msgSplits[2]] = true

			//setting slave free again to search
			for k := range slaves {
				if slaves[k].conn == c {
					t := slaves[k]
					t.currentSearchFile = ""
					t.freeToSearch = true
					slaves[k] = t
				}
			}

			// fmt.Println("slaves", slaves)

			passwordNotFound(msgSplits[1])

		} else if msgSplits[0] == "hb" {

			for k := range slaves {
				if slaves[k].conn == c {
					fmt.Println("Slave Alive...")
					t := slaves[k]
					t.alive = true
					slaves[k] = t
				}
			}
		} else if msgSplits[0] == "f" { //this means that server specifically asked for files again and this only occurs when refreshing the slaveFiles map
			filesArray := strings.Split(msgSplits[1], ",")

			for _, f := range filesArray {

				//set new map value if only not already set
				if _, ok := slaveFiles[f]; !ok {
					slaveFiles[f] = false
				}

			}
			fmt.Println(slaveFiles)
		}

		// defer func() {
		// 	log.Printf("Connection from Slave %v closed.\n", c.RemoteAddr())
		// 	fmt.Println("slaveFiles", slaveFiles)
		//
		//  deleteSlave(c)
		//
		// 	fmt.Println("slaves", slaves)
		//
		// }()

	}
}

//deletes slave from map, removes files related to it and reload slaveFiles from all connected slaves
func deleteSlave(index int) {

	slaveToBeDeleted := slaves[index]

	// Remove slave from slaves array
	slaves[index] = slaves[len(slaves)-1] // Copy last element to index i.
	// slaves[len(slaves)-1] = nil           // Erase last element (write zero value).
	slaves = slaves[:len(slaves)-1] // Truncate slice.

	filesToBeDeleted := slaveToBeDeleted.files

	for k := range slaveFiles {
		for _, f := range filesToBeDeleted {
			if k == f {
				fmt.Println("File deleted", k)
				delete(slaveFiles, k)
			}
		}
	}
	fmt.Println("Slave deleted", slaveToBeDeleted)

	//ask all slaves to send respective files so that slaveFiles Map can be updated
	for _, v := range slaves {
		v.conn.Write([]byte("sf"))
	}
}

func checkAliveSlaves() {
	//repeat process periodically
	ticker := time.NewTicker(15 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				for k := range slaves {
					if !slaves[k].alive {
						// if slaves[k].currentSearchFile != "" {
						//
						// 	slaveFiles[slaves[k].currentSearchFile] = false
						//
						// }
						deleteSlave(k)

						fmt.Println("Dead slave deleted.")
					}
				}
				//Set all other slaves to be tested again if alive or not
				for k := range slaves {

					slaves[k].alive = false

				}
				fmt.Println("slaveFiles::::", slaveFiles)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func handleSlaves(myIP string, slavePort string) {

	ln, err := net.Listen("tcp", myIP+":"+slavePort)
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

func handlePendingClients() {
	if len(clients) > 0 {
		fmt.Println("Handling pending client requests")
		sendPasswordToSlaves(clients[0])
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

		var slaveFilesCopy = make(map[string]bool)

		for k, v := range slaveFiles {
			slaveFilesCopy[k] = v
		}

		currentClient := Client{password, c, slaveFilesCopy}
		//add to array
		clients = append(clients, currentClient)

		go sendPasswordToSlaves(currentClient)

		defer func() {
			fmt.Println("Connection from Client closed.\n")

			for k, v := range clients {
				if v.conn == c {
					clients[k] = clients[len(clients)-1]
					clients = clients[:len(clients)-1]
					break
				}
			}
			fmt.Println("Client deleted from array.")

			go handlePendingClients()
		}()
	}

}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func main() {

	myIP := GetOutboundIP()
	var clientPort string
	var slavePort string
	var localClient bool
	flag.StringVar(&clientPort, "clientPort", "8000", "Port on which server will listen for client connection.")
	flag.StringVar(&slavePort, "slavePort", "8001", "Port on which server will listen for slave connection.")
	flag.BoolVar(&localClient, "localClient", true, "True if client is local, false otherwise.")

	flag.Parse()

	go handleSlaves(myIP, slavePort)

	// go checkAliveSlaves()

	if localClient {
		fmt.Println("Running server...\nListening for Clients on localhost:" + clientPort + "\nListening for Slaves on " + myIP + ":" + slavePort)

		ln, err := net.Listen("tcp", "127.0.0.1:"+clientPort)
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
	} else {
		fmt.Println("Running server...\nListening for Clients on " + myIP + ":" + clientPort + "\nListening for Slaves on " + myIP + ":" + slavePort)

		ln, err := net.Listen("tcp", myIP+":"+clientPort)
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

}
