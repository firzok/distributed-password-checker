// package main
//
// import (
// 	"bufio"
// 	"fmt"
// 	"math/rand"
// 	"net"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"time"
// )
//
// func handleConnection(c net.Conn) {
// 	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
// 	for {
// 		netData, err := bufio.NewReader(c).ReadString('\n')
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
//
// 		temp := strings.TrimSpace(string(netData))
// 		if temp == "STOP" {
// 			break
// 		}
//
// 		rand.Seed(time.Now().UnixNano())
// 		min := 10
// 		max := 30
// 		result := strconv.Itoa(rand.Intn(max-min)+min) + "\n"
// 		c.Write([]byte(string(result)))
// 	}
// 	c.Close()
// }
//
// func main() {
// 	arguments := os.Args
// 	if len(arguments) == 1 {
// 		fmt.Println("Please provide a port number!")
// 		return
// 	}
//
// 	PORT := ":" + arguments[1]
// 	l, err := net.Listen("tcp4", PORT)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer l.Close()
// 	rand.Seed(time.Now().Unix())
//
// 	for {
// 		c, err := l.Accept()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		go handleConnection(c)
// 	}
// }
package main

import (
	"net/http"
)

// func (clientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Listening on 8001: Client Port "))
// }
// func (m *slaveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("Listening on 8002: Slave Port"))
// }

// func clientHandler(w http.ResponseWriter, r *http.Request) {
// 	message := r.URL.Path
// 	message = strings.TrimPrefix(message, "/")
// 	message = "Hello " + message
// 	w.Write([]byte(message))
// }
// func slaveHandler(w http.ResponseWriter, r *http.Request) {
// 	message := r.URL.Path
// 	message = strings.TrimPrefix(message, "/")
// 	message = "Hello " + message
// 	w.Write([]byte(message))
// }
//
// func main() {
// 	arguments := os.Args
//
// 	if len(arguments) == 1 {
// 		fmt.Println("Please Provide a Port number")
// 		return
// 	}
//
// 	http.HandleFunc("/", clientHandler)
// 	http.HandleFunc("/", slaveHandler)
// 	clientPORT := ":" + arguments[1]
// 	slavePORT := ":" + arguments[2]
//
// 	go http.ListenAndServe(clientPORT, nil)
// 	go http.ListenAndServe(slavePORT, nil)
//
// 	// if err := http.ListenAndServe(":8080", nil); err != nil {
// 	// 	panic(err)
// 	// }
// }

func main() {
	go func() {
		http.ListenAndServe(":8001", &clientHandler{})
	}()

	//the last call is outside goroutine to avoid that program just exit
	http.ListenAndServe(":8002", &slaveHandler{})
}

type clientHandler struct {
}

func (m *clientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Listening on 8001: client "))
}

type slaveHandler struct {
}

func (m *slaveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Listening on 8002: slave "))
}
