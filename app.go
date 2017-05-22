package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	urlPrefix  = "/app/"
	defaultImg = "/home/gideon/work/imgsrvr/testingpics/Graphic1.jpg"
	imgHash    = 6
)

type FastCGIServer struct{}
type tData struct {
	Fn string
	Tn string
}

//First page Stuff!
func appPage(resp http.ResponseWriter, req *http.Request) {
	firstPageTemplate := template.New("first page templated.")
	firstPageTemplate, err := firstPageTemplate.Parse(firstPage)
	if err != nil {
		fmt.Printf("Failed to parse template: %v", err)
	}
	field := req.FormValue("fn")
	fmt.Println(field)
	tData := tData{
		Tn: field,
	}
	if err = firstPageTemplate.Execute(resp, tData); err != nil {
		fmt.Printf("template execute error: %v", err)
	}
}

//testingPage!!!
func testingPage(resp http.ResponseWriter, req *http.Request) {
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

// Page for sending pics
func sendImg(resp http.ResponseWriter, req *http.Request, img string) {
	if len(img) != imgHash {
		img = defaultImg
	}
	//Check if file exists and open
	openfile, err := os.Open(img)
	defer openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(resp, "File not found.", 404)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	openfile.Read(fileHeader)
	//Get content type of file
	fileContentType := http.DetectContentType(fileHeader)
	//Get the file size
	fileStat, _ := openfile.Stat()                     //Get info from file
	fileSize := strconv.FormatInt(fileStat.Size(), 10) //Get file size as a string

	//Send the headers
	//resp.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	resp.Header().Set("Content-Type", fileContentType)
	resp.Header().Set("Content-Length", fileSize)

	//Send the file
	//We read 512 bytes from the file already so we reset the offset back to 0
	openfile.Seek(0, 0)
	io.Copy(resp, openfile) //'Copy' the file to the client
	return
}

func errorHandler(resp http.ResponseWriter, req *http.Request, status int) {
	resp.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(resp, "custom 404")
	}
}

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("the req arrived")
	if req.Body == nil {
		return
	}
	fmt.Printf("URL is: %v\n", req.URL.Path)

	/*TODO: Quality check URL before switch
	Quality check = make sure the URL wont kill the script on the switch statement.
	Things to look for: How many "elements" are in the URL
	make sure it fits the pattern (/app/foo)

	*/
	urlSplit := strings.Split(req.URL.Path, "/")
	urlECount := len(urlSplit)
	fmt.Println(urlECount)
	if urlECount < 3 {
		errorHandler(resp, req, http.StatusNotFound)
	}
	if !strings.HasPrefix(req.URL.Path, urlPrefix) {
		errorHandler(resp, req, http.StatusNotFound)
	}

	// Now URL looks like "urlPrefix/foo"
	switch urlSplit[2] {
	case "test":
		testingPage(resp, req)
	case "i":
		if urlECount != 4 {
			errorHandler(resp, req, http.StatusNotFound)
		}
		sendImg(resp, req, urlSplit[3])
	default:
		appPage(resp, req)
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

	//Debug:
	//This prints stuff in the console so i get info, just for me
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error happened!!! Here, take it: %v", err)
	}
	fmt.Printf("DIR: %v\n", dir)
	//end of Debug

	fcgi.Serve(listener, srv) //end of request
}
