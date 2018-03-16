package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	urlPrefix  = "/app/"
	defaultImg = "/home/gido5731/work/imgsrvr/testingpics/Graphic1-50.jpg"
	imgHash    = 6
	imgStore   = "/var/tmp/imgStorage/"
)

type FastCGIServer struct{}
type tData struct {
	Fn string
	Tn string
}

//404 support, I dont know why I did this but I am too scared to undo it at this point
func notFound(resp http.ResponseWriter, req *http.Request, status int) {
	resp.WriteHeader(status)
	fmt.Fprint(resp, "custom ", status)
}

//First page Stuff!
func appPage(resp http.ResponseWriter, req *http.Request) {
	firstPageTemplate := template.New("first page templated.")
	firstPageTemplate, err := firstPageTemplate.Parse(firstPage)
	if err != nil {
		fmt.Printf("Failed to parse template: %v", err)
		return
	}

	field := req.FormValue("fn")
	fmt.Println(field)
	tData := tData{
		Fn: field,
	}
	//upload(resp, req)
	fmt.Printf("Form data: ", field, "\ntData: ", tData)

	if err = firstPageTemplate.Execute(resp, tData); err != nil {
		fmt.Printf("template execute error: %v", err)
		return

	}

}

//testingPage!!! Missleading name, I know, this page takes the info from last page to upload
func testingPage(resp http.ResponseWriter, req *http.Request) {
	/* TODO:
	store file on disk:
	-Accept the file 								DONE
	-create name (from md5)							DONE
	-create a map/index of pub name (hash) to path	(Postponed)
	provide path to file
	*/
	http.HandleFunc("/img", upload)
	upload(resp, req)

}
func upload(resp http.ResponseWriter, req *http.Request) {

	fmt.Println("method:", req.Method)
	if req.Method == "GET" {
		crutime := time.Now().Unix()
		fmt.Println("Beep Beep Beep... The time is:", crutime)
		md5 := md5.New()
		io.WriteString(md5, strconv.FormatInt(crutime, 10))
		encodedMd5 := make([]byte, hex.EncodedLen(imgHash))
		byteMd5 := []byte("md5")
		hex.Encode(byteMd5, encodedMd5)
		fmt.Println("I just got the hashed md5! Here it is:", encodedMd5, "\nEnd of md5")
		//fmt.Printf("MD5:", md5)
		token := fmt.Sprintf("%x", md5.Sum(nil))
		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(resp, token)
	} else {
		req.ParseMultipartForm(32 << 20)
		crutime := time.Now().Unix()
		fmt.Println("Beep Beep Beep... The time is:", crutime)

		file, handler, err := req.FormFile("img")
		if err != nil {
			fmt.Println(err)
			return
		}
		md5 := md5.New()
		io.WriteString(md5, strconv.FormatInt(crutime, 10))
		byteMd5 := []byte("md5")
		encodedMd5 := hex.EncodeToString(byteMd5)
		fmt.Println("I just got the hashed md5! Here it is:", encodedMd5, "\nEnd of md5")
		//fmt.Printf("MD5:", md5)
		//token := fmt.Sprintf("%x", md5.Sum(nil))
		//t, _ := template.ParseFiles("upload.gtpl")
		//t.Execute(resp, token)
		fmt.Println("file: ", file)
		firstChar := string(encodedMd5[0])
		secondChar := string(encodedMd5[1])
		//fmt.Printf("File:", file, "\nhandler: ", handler) //too spammy for normal use
		defer file.Close()
		fmt.Fprintf(resp, "%v", handler.Header)
		filepath := imgStore + firstChar + "/" + secondChar + "/"
		f, err := os.OpenFile(filepath+encodedMd5, os.O_WRONLY|os.O_CREATE, 0666)
		fmt.Println("filename?: ", filepath+encodedMd5)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

// Page for sending pics
func sendImg(resp http.ResponseWriter, req *http.Request, img string) {
	fmt.Println("Recieved a req to send the user a file")
	if len(img) != imgHash {
		//img = defaultImg //if no image exists, use testing image
		//fmt.Println("Using Default Image")
		errorHandler(resp, req, 404)
	}
	firstChar := string(img[0])
	secondChar := string(img[1])
	filepath := imgStore + firstChar + "/" + secondChar + "/"
	//Check if file exists and open
	openfile, err := os.Open(filepath + img)
	defer openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		errorHandler(resp, req, http.StatusNotFound)
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
	fmt.Printf("Heres the file size: ", fileSize)

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
		//fmt.Fprint(resp, "custom 404")
		notFound(resp, req, status)
	}
}
func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	fmt.Println("the req arrived")
	if req.Body == nil {
		return
	}
	fmt.Printf("URL is: %v\n", req.URL.Path)

	urlSplit := strings.Split(req.URL.Path, "/")
	urlECount := len(urlSplit)
	fmt.Printf("urlECount: %d\n", urlECount)
	// Checking amt of elements in url (else sends 404)
	if urlECount < 3 {
		errorHandler(resp, req, http.StatusNotFound)
		return
	}
	// Check for prefix
	if !strings.HasPrefix(req.URL.Path, urlPrefix) {
		errorHandler(resp, req, http.StatusNotFound)
		return
	}

	// Now URL looks like "urlPrefix/foo"
	switch urlSplit[2] {
	case "test":
		testingPage(resp, req)
	case "i":
		// Checks for hash/element/thing
		if urlECount != 4 || urlSplit[3] == "" {
			errorHandler(resp, req, http.StatusNotFound)
			return
		}
		fmt.Printf("urlECount of IMG: %d\n", urlECount)
		fmt.Printf("Split for image: %v\n", urlSplit)
		sendImg(resp, req, urlSplit[3])
		//upload(resp, req)
	default:
		appPage(resp, req)
	}
}

// createImgDir creates all image storage directories
func createImgDir(imgStore string) {
	for f := 0; f < 16; f++ {
		for s := 0; s < 16; s++ {
			os.MkdirAll(filepath.Join(imgStore, fmt.Sprintf("%x/%x", f, s)), 0755)
		}
	}
	fmt.Println("Finished making/Verifying the directories!")
}

//When everything gets set up, all page setup above this
func main() {

	go createImgDir(imgStore)

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

//LEGACY CODE (Only here for historical purposes, please ignore)
/*
//Handling Uploading, will one day put into a func of its own, one day.... (this was in testingPage)
	req.ParseMultipartForm(32 << 20)
	file, handler, err := req.FormFile("img")
	if err != nil {
		fmt.Println(err)
		return //checks for file
	}
	defer file.Close()
	fmt.Fprintf(resp, "%v", handler.Header)
	f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	// filename := fileHash
	testPageTemplate := template.New("test page templated.")
	testPageTemplate, err = testPageTemplate.Parse(testPage)
	if err != nil {
		fmt.Printf("Failed to parse template: %v", err) // this only happens if someone goofs the template file
		return
	}
	field := req.FormValue("tn")
	fmt.Println(field)
	tData := tData{
		Tn: field,
	}
	if err = testPageTemplate.Execute(resp, tData); err != nil {
		fmt.Printf("template execute error: %v", err)
		return
	}


//FileSplit (From upload)
	//fileSplit := strings.Split(imgStore+handler.Filename, "/")
	//fmt.Println("filesplit: ", fileSplit[4])
*/
