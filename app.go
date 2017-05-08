package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/fcgi"

	_ "github.com/go-sql-driver/mysql"
)

type FastCGIServer struct{}
type tData struct {
	Fn string
	Tn string
}

func appPage(resp http.ResponseWriter, req *http.Request) {
	//First page Stuff!
	firstPageTemplate := template.New("first page templated.")
	firstPageTemplate, err := firstPageTemplate.Parse(firstPage)
	if err != nil {
		fmt.Printf("Failed to parse template: %v", err)
	}
	field := req.FormValue("fn")
	fmt.Println(field)
	tData := tData{
		Fn: field,
	}
	if err = firstPageTemplate.Execute(resp, tData); err != nil {
		fmt.Printf("template execute error: %v", err)
	}
}

func testingPage(resp http.ResponseWriter, req *http.Request) {
	//testingPage!!!
	testPageTemplate := template.New("test page templated.")
	testPageTemplate, err := testPageTemplate.Parse(testPage)
	if err != nil {
		fmt.Printf("Failed to parse template: %v", err)
	}
	field := req.FormValue("tn")
	fmt.Println(field)
	tData := tData{
		Tn: field,
	}
	if err = testPageTemplate.Execute(resp, tData); err != nil {
		fmt.Printf("template execute error: %v", err)
	}
}

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("the req arrived")
	if req.Body == nil {
		return
	}
	fmt.Printf("URL is: %v\n", req.URL.Path)

	switch req.URL.Path {
	default:
		appPage(resp, req)
	/*case "/app/main/":
	testingPage(resp, req)*/
	case "/app/test/":
		testingPage(resp, req)
	case "app/test":
		testingPage(resp, req)
	}
}

//When everything gets set up, all page setup above this
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
