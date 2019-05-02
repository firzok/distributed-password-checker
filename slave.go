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

func searchPasswordInFile(password string, file string) string {
	f, err := os.Open(file)
	if err != nil {
	}
	defer f.Close()

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), password) {
			return "1"
		}
	}
	return "0"
}

func performSlaveoperations(c net.Conn, newsearchchan <-chan SearchQuery) {
	for {
		select {
		case search := <-newsearchchan:
			log.Printf("New Search: %s in %s", search.password, search.fileName)
			ret := searchPasswordInFile(search.password, search.fileName)
			//send result to server (either found or not found)
			c.Write([]byte(ret))
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

	go performSlaveoperations(conn, searchchan)

	handleSlaveOperations(conn, searchchan)

}
