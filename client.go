package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
)

func getPassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		fmt.Println("password:", r.Form["password"])

		sendPasswordToServer(r.Form["password"][0])

	}
}

func sendPasswordToServer(password string) {
	conn, err := net.Dial("tcp", "127.0.0.1:8003")
	if err != nil {
		fmt.Println("ERROR")
		return
	}
	defer conn.Close()
	conn.Write([]byte(password))
}

func main() {

	http.HandleFunc("/", getPassword)         // setting router rule
	err1 := http.ListenAndServe(":9095", nil) // setting listening port
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}
}
