package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

type SearchQuery struct {
	password string
	fileName string
}

func searchPasswordInFile(password string, file string, stopSearchChan chan SearchQuery) int {
	f, err := os.Open("./passwordSplitFiles/" + file)
	if err != nil {
		fmt.Println("Error opening file")
	}
	defer f.Close()

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		if scanner.Text() == password {
			return 1
		}
		if i%500 == 0 {
			select {
			case stopSearch := <-stopSearchChan:
				if stopSearch.password == password {
					break
				}

			}

		}

		i++
	}
	return 0
}

func performSlaveoperations(c net.Conn, newsearchchan chan SearchQuery, stopSearchChan chan SearchQuery) {
	for {
		select {
		case search := <-newsearchchan:
			log.Printf("New Search: %s in %s", search.password, search.fileName)
			ret := searchPasswordInFile(search.password, search.fileName, stopSearchChan)

			//send result to server (either found or not found)
			if ret == 1 {
				fmt.Println("Password Found")
				c.Write([]byte("pf:" + search.password + ":" + search.fileName))
			} else {
				fmt.Println("Password NOT Found")
				c.Write([]byte("pnf:" + search.password + ":" + search.fileName))
			}

		}
	}
}

func handleSlaveOperations(c net.Conn, searchchan chan SearchQuery) {
	buf := make([]byte, 4096)
	defer c.Close()

	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		command := strings.Split(string(buf[0:n]), ":")

		if command[0] == "s" {
			search := SearchQuery{command[1], command[2]}
			searchchan <- search

		} else if command[0] == "pf" {

		}

	}
}

func main() {

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

	fmt.Println(fileNames)

	conn, err := net.Dial("tcp", "127.0.0.1:8002")
	if err != nil {
		fmt.Println("ERROR: Connecting to Server")
		return
	}
	conn.Write([]byte(fileNames))

	searchchan := make(chan SearchQuery)
	stopSearchChan := make(chan SearchQuery)

	go performSlaveoperations(conn, searchchan, stopSearchChan)

	handleSlaveOperations(conn, searchchan)

}
