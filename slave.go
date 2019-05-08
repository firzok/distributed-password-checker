package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type SearchQuery struct {
	password string
	fileName string
}

func sendHeartBeats(c net.Conn, quit chan struct{}) {

	//send heartbeats every 5 seconds
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:

				c.Write([]byte("hb"))

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func searchPasswordInFile(password string, file string, stopSearchChan chan string) int {

	f, err := os.Open("./passwordSplitFiles/" + file)
	if err != nil {
		fmt.Println("Error opening file")
		return 0
	}
	defer f.Close()

	fmt.Println("Searching password " + password + "in file " + file)

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		select {
		//Some other slave found the password so break loop
		case stopSearchForPassword := <-stopSearchChan:
			log.Println("Stop search request")
			if stopSearchForPassword == password {
				fmt.Println("Search STOPPED")
				return 2
			}
		default:
			//check for password
			// fmt.Println(scanner.Text())
			if scanner.Text() == password {
				return 1
			}
		}

	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return 0
}

func performSlaveoperations(c net.Conn, newsearchchan chan SearchQuery, stopSearchChan chan string) {
	for {
		select {
		case search := <-newsearchchan:
			log.Printf("New Search: " + search.password + " in " + search.fileName)
			ret := searchPasswordInFile(search.password, search.fileName, stopSearchChan)

			//send result to server (either found or not found)
			if ret == 1 {
				fmt.Println("Password Found")
				c.Write([]byte("pf:" + search.password + ":" + search.fileName))
			} else if ret == 0 {
				fmt.Println("Password NOT Found")
				fmt.Println(search.fileName)
				c.Write([]byte("pnf:" + search.password + ":" + search.fileName))
			} else if ret == 2 {
				fmt.Println("Password Found By Some Other Slave")
			}

		}
	}
}

func handleSlaveOperations(c net.Conn, searchchan chan SearchQuery, stopSearchChan chan string) {

	defer c.Close()

	for {
		buf := make([]byte, 4096)
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		command := strings.Split(string(buf[0:n]), ":")
		fmt.Println(command)
		if command[0] == "s" { //search
			search := SearchQuery{command[1], command[2]}
			searchchan <- search

		} else if command[0] == "pf" { //password found
			stopSearchChan <- command[1]
		} else if command[0] == "sf" { //send files
			fileNames := getFileNames()
			fmt.Println("Files: ", fileNames)
			c.Write([]byte("f:" + fileNames))

		}

	}
}

func getFileNames() string {
	files, err := ioutil.ReadDir("./passwordSplitFiles/")
	if err != nil {
		log.Fatal(err)
	}
	var fileNames string
	for i, file := range files {
		fileNames += file.Name()

		if i != len(files)-1 {
			fileNames += ","
		}
	}

	return fileNames
}

func main() {
	var serverPort string
	var serverIP string

	flag.StringVar(&serverPort, "serverPort", "8001", "Port on which slave will connect to server.")
	flag.StringVar(&serverIP, "serverIP", "127.0.0.1", "IP on which slave will connect to server.")
	flag.Parse()

	fileNames := getFileNames()

	fmt.Println("Files: ", fileNames)
	fmt.Println("Connecting to Server on Port: " + serverPort)

	conn, err := net.Dial("tcp", serverIP+":"+serverPort)
	if err != nil {
		fmt.Println("ERROR: Connecting to Server")
		return
	}
	conn.Write([]byte(fileNames))

	searchchan := make(chan SearchQuery)
	stopSearchChan := make(chan string)

	go performSlaveoperations(conn, searchchan, stopSearchChan)

	stopHeartBeat := make(chan struct{})

	go sendHeartBeats(conn, stopHeartBeat)

	handleSlaveOperations(conn, searchchan, stopSearchChan)

}
