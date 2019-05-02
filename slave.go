package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

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
	conn.Close()
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
