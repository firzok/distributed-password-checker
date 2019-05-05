//Kill a process on port
//kill -kill $(lsof -t -i :9090)
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
)

var pwned = make(chan bool, 1)

// func wait(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("wait", r.Method)
//
// 	if r.Method == "POST" {
//
// 		r.ParseForm()
// 		password := r.Form["password"][0]
//
// 		http.ServeFile(w, r, "wait.html")
//
// 		fmt.Println("password:", password)
// 		result := sendPasswordToServer(password)
//
// 		if result == "pf" {
// 			fmt.Println("You password has been PWNED1.")
// 			pwned <- true
//
// 		} else {
// 			fmt.Println("You password is secure1.")
// 			pwned <- false
// 		}
//
// 	} else {
// 		select {
// 		default:
// 			http.ServeFile(w, r, "wait.html")
// 		case t := <-pwned:
//
// 			if t {
// 				http.ServeFile(w, r, "pwned.html")
// 			} else {
// 				http.ServeFile(w, r, "secure.html")
// 			}
//
// 		}
// 	}
//
// }

func getPassword(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {

		select {
		default:
			http.ServeFile(w, r, "login.html")

		case t := <-pwned:

			if t {
				http.ServeFile(w, r, "pwned.html")
			} else {
				http.ServeFile(w, r, "secure.html")
			}

		}

	} else if r.Method == "POST" {

		r.ParseForm()
		password := r.Form["password"][0]
		fmt.Println("password:", password)

		// w.Write([]byte("Hang on tight while we are looking for your Password in our Database..."))
		go func() {
			http.ServeFile(w, r, "wait.html")
		}()

		result := sendPasswordToServer(password)

		if result == "pf" {
			fmt.Println("You password has been PWNED.")
			pwned <- true

		} else {
			fmt.Println("You password is secure.")
			pwned <- false

		}

	}
}

func sendPasswordToServer(password string) string {
	conn, err := net.Dial("tcp", "127.0.0.1:"+serverPort)
	if err != nil {
		fmt.Println("ERROR")
		return "pf"
	}
	defer conn.Close()
	conn.Write([]byte(password))

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		fmt.Println("ERROR: Getting result from server.")
	}

	result := string(buf[0:n])
	fmt.Println(result)

	return result

}

var serverPort string

func main() {
	var clientPort string

	flag.StringVar(&clientPort, "clientPort", "9090", "Port on which client will run on localhost.")
	flag.StringVar(&serverPort, "serverPort", "8000", "Port on which client will connect to server.")

	flag.Parse()

	fmt.Println("Running Client on 127.0.0.1:" + clientPort)
	// http.HandleFunc("/wait", wait)
	http.HandleFunc("/pwned", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pwned.html")
	})

	http.HandleFunc("/", getPassword)                // setting password getting function
	err1 := http.ListenAndServe(":"+clientPort, nil) // setting port

	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}

}
