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

		result := sendPasswordToServer(r.Form["password"][0])
		if result == "pf" {
			w.Write([]byte("You password has been PWNED."))
		} else {
			w.Write([]byte("You password is secure."))
		}

	}
}

func sendPasswordToServer(password string) string {
	conn, err := net.Dial("tcp", "127.0.0.1:8003")
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

func main() {

	http.HandleFunc("/", getPassword)         // setting router rule
	err1 := http.ListenAndServe(":9095", nil) // setting listening port
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}
}
