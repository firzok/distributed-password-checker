package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

type SearchQuery struct {
	password string
	fileName string
}

func performSlaveoperations(searchchan <-chan SearchQuery) {
	for {
		select {
		case search := <-searchchan:
			log.Printf("New Search: %s in %s", search.password, search.fileName)

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

	go performSlaveoperations(searchchan)

	handleSlaveOperations(conn, searchchan)

	// f, err := os.Open("TestPage.txt")
	// if err != nil {
	// }
	// defer f.Close()
	//
	// // Splits on newlines by default.
	// scanner := bufio.NewScanner(f)
	// for scanner.Scan() {
	// 	if strings.Contains(scanner.Text(), "sample") {
	// 		//fmt.Print("Found!")
	// 	}
	// }
	// myfile, _ := ioutil.ReadFile("TestPage.txt")
	// if strings.Contains(string(myfile), "sample") {
	// 	fmt.Print("Found!")
	// }
}
