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

func stringInSlice(a string, list []string) bool {
	//Searches for a string in list of strings and return true or false
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func sendPasswordToSlaves(client Client) {

	password := client.password

	//loop over all slaves
	for slaveIndex, slave := range slaves {

		//if that slave is free to search
		if slave.freeToSearch {

			//loop over all the Slave files
			for fileName, searched := range client.searchFiles {

				//if that file has not yet been searched and the slave has that file then search that file for password
				if !searched && stringInSlice(fileName, slave.files) {
					//sending command to search for password in that file
					slave.conn.Write([]byte("s:" + password + ":" + fileName))
					fmt.Println("Password: " + password + " sent to slave to be searched in " + fileName)

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

	//Sedning result to client who asked for this password
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

}

func passwordNotFound(password string) {

	if len(clients) < 1 {
		return
	}

	var currentClientIndex = 0
	for i, client := range clients {
		if client.password == password {
			currentClientIndex = i
			break
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

	//all files have been searched and still password not Found then
	//loop over all clients
	for _, c := range clients {

		//if client asked for this password
		if c.password == password {

			//tell client that password not found
			c.conn.Write([]byte("pnf"))
			fmt.Println("Result sent to Client")
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
	//make new slave object
	currentSlave := Slave{filesArray, c, true, true, ""}
	//add to the array of slaves
	slaves = append(slaves, currentSlave)

	fmt.Println("Files from new Slave: ", filesArray)

	//add filesArray to the bigger list of slavefiles Map
	for _, f := range filesArray {
		if _, ok := slaveFiles[f]; !ok {
			slaveFiles[f] = false
		}

	}

	// fmt.Println("slaveFiles", slaveFiles)

	for { //Main slave handler loop
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
			fmt.Println("Password::::" + msgSplits[1] + " Found in:" + msgSplits[2])
			passwordFoundBySlave(msgSplits[1], c)
		} else if msgSplits[0] == "pnf" { //password not found
			fmt.Println("Password::::" + msgSplits[1] + " NOT Found in:" + msgSplits[2])

			//Setting file searched to be true so that no other slave searches this file
			// if len(clients) > 0 {
			// 	var currentClientIndex = 0
			// 	for i, client := range clients {
			// 		if client.password == msgSplits[1] {
			// 			currentClientIndex = i
			// 			break
			// 		}
			// 	}
			// 	clients[currentClientIndex].searchFiles[msgSplits[2]] = true
			// }

			//setting slave free again to search
			for k := range slaves {
				if slaves[k].conn == c {
					slaves[k].currentSearchFile = ""
					slaves[k].freeToSearch = true
				}
			}
			//send password again to be searched in some other file
			passwordNotFound(msgSplits[1])

		} else if msgSplits[0] == "hb" {

			for k := range slaves {
				if slaves[k].conn == c {
					fmt.Println("Heartbeat::::Slave is alive...")
					slaves[k].alive = true

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
			// fmt.Println(slaveFiles)
		}

		//Another way of handling disconnected slaves
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
	slaves = slaves[:len(slaves)-1]       // Truncate slice.

	filesToBeDeleted := slaveToBeDeleted.files

	for k := range slaveFiles {
		for _, f := range filesToBeDeleted {
			if k == f {
				fmt.Println("File deleted from slaveFiles", k)
				delete(slaveFiles, k)
			}
		}
	}

	//If slave was searching some file while it got disconnected
	if slaveToBeDeleted.currentSearchFile != "" {
		//loop over all clients
		for _, v := range clients {
			//if that client has that file in its searchFiles map
			if _, ok := v.searchFiles[slaveToBeDeleted.currentSearchFile]; ok {
				//make it false so that it gets searched again by some other slave
				v.searchFiles[slaveToBeDeleted.currentSearchFile] = false
			}
		}
	}

	fmt.Println("Slave deleted:", slaveToBeDeleted)

	//ask all slaves to send respective files so that slaveFiles Map can be updated
	//updated in main loop
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
				for k, v := range slaves {
					if !v.alive {
						//if slave is not alive delete it
						//This only happens if slave failed to send a Heartbeat in the last period of time
						deleteSlave(k)

						fmt.Println(":::Dead slave deleted:::\n")
					}
				}
				//Set all other slaves.alive to false to be tested again if alive or not
				for k := range slaves {
					slaves[k].alive = false
				}

				// fmt.Println("slaveFiles::::", slaveFiles)
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
		fmt.Println("\nERROR::Unable to create Slave Server.")
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
		fmt.Println("Handling pending client requests...")
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

			//Remove from client array
			for k, v := range clients {
				if v.conn == c {
					clients[k] = clients[len(clients)-1]
					clients = clients[:len(clients)-1]
					break
				}
			}
			fmt.Println("Client deleted from array.")
			if len(clients) > 0 {
				go handlePendingClients()
			}
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

	// Check for alive slaves after regular intervals
	go checkAliveSlaves()

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
