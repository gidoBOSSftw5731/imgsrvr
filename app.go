package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/fcgi"

	_ "github.com/go-sql-driver/mysql"
)

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("the req arrived")
	if req.Body == nil {
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	//Getting stuff from a GET!! (or to normies, putting together stuff to send)
	fieldValue := req.FormValue("field_name")
	ret := fmt.Sprintf("<h1>Hello, 世界</h1>\n<p>Behold my Go web app.</p> %s", fieldValue)
	type UserInput struct {
		SomeField    string
		AnotherField int
		LastOne      string
	}
	var inp UserInput
	json.Unmarshal(body, &inp)
	SomeField := "test"
	resp.Write([]byte(m.SomeField))
	// now write out data!
	resp.Write([]byte(ret))
}

func main() {
	fmt.Println("Starting the program.")
	listener, _ := net.Listen("tcp", "127.0.0.1:9001")
	fmt.Println("Started the listener.")
	srv := new(FastCGIServer)
	fmt.Println("Starting the fcgi.")

	conn, err := sql.Open("mysql", "/test")
	fmt.Println("Oi, mysql did thing")
	defer conn.Close()

	if err != nil {
		fmt.Println("Oh noez, could not connect to database")
		return
	}

	fcgi.Serve(listener, srv) //end of request
}
