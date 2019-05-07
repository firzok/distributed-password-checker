//Kill a process on port
//kill -kill $(lsof -t -i :9090)
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
)

var page = make(chan string, 1)

type ResultPage struct {
	Result   string
	Redirect string
}

func wait(w http.ResponseWriter, r *http.Request) {
	fmt.Println("wait:", r.Method)
	tmpl, err := template.ParseFiles("wait.html")

	if err != nil {
		fmt.Println("Error Parsing file.")
	}
	if r.Method == "GET" {

		data := ResultPage{
			Result: "Hang on tight while we are looking for your Password in our Database..."}
		select {
		default:

		case t := <-page:

			if t == "pwned" {
				http.Redirect(w, r, "http://127.0.0.1:"+clientPort+"/pwned/", http.StatusSeeOther)
				// data = ResultPage{
				// Result: "You password has been PWNED.\n\nGo back to search new Password."}
			} else if t == "secure" {
				http.Redirect(w, r, "http://127.0.0.1:"+clientPort+"/secure/", http.StatusSeeOther)

				// data = ResultPage{
				// Result: "You password is SECURE.\n\nGo back to search new Password."}
			}

		}

		tmpl.Execute(w, data)
	} else if r.Method == "POST" {

		r.ParseForm()
		password := r.Form["password"][0]
		fmt.Println("password:", password)

		data := ResultPage{
			Result: "Hang on tight while we are looking for your Password in our Database..."}
		tmpl.Execute(w, data)

		go func() {
			result := sendPasswordToServer(password)

			if result == "pf" {
				fmt.Println("You password has been PWNED.")
				page <- "pwned"

			} else {
				fmt.Println("You password is secure.")
				page <- "secure"

			}
		}()

	}

}

func getPassword(w http.ResponseWriter, r *http.Request) {

	fmt.Println("getPassword:", r.Method) //get request method

	http.ServeFile(w, r, "password.html")

}

func sendPasswordToServer(password string) string {
	conn, err := net.Dial("tcp", serverIP+":"+serverPort)
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
var serverIP string
var clientPort string

func main() {
	page <- "login"

	flag.StringVar(&clientPort, "clientPort", "9090", "Port on which client will run on localhost.")
	flag.StringVar(&serverPort, "serverPort", "8000", "Port on which client will connect to server.")
	flag.StringVar(&serverIP, "serverIP", "127.0.0.1", "IP on which client will connect to server.")

	flag.Parse()

	fmt.Println("Running Client on 127.0.0.1:" + clientPort)
	http.HandleFunc("/wait/", wait)
	http.HandleFunc("/pwned/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("result.html")

		if err != nil {
			fmt.Println("Error Parsing file.")
		}
		data := ResultPage{
			Result:   "You password has been PWNED...",
			Redirect: "/pass/"}

		tmpl.Execute(w, data)
	})

	http.HandleFunc("/secure/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("result.html")

		if err != nil {
			fmt.Println("Error Parsing file.")
		}
		data := ResultPage{
			Result:   "You password is SECURE...",
			Redirect: "/pass/"}

		tmpl.Execute(w, data)
	})

	http.HandleFunc("/", getPassword) // setting password getting function
	http.HandleFunc("/pass/", getPassword)
	err1 := http.ListenAndServe(":"+clientPort, nil) // setting port

	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}

}
