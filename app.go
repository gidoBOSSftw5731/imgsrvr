package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type FastCGIServer struct{}
type tData struct {
	Fn string
	Tn string
}

type FileHeader struct {
	Filename string
	Header   textproto.MIMEHeader
	// contains filtered or unexported fields
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
	req.ParseForm()
	field := req.FormValue("fn")
	fmt.Println(field)
	tData := tData{
		Fn: field,
	}
	//upload(resp, req)
	fmt.Println("Form data: ", field, "\ntData: ", tData)
	if err = firstPageTemplate.Execute(resp, tData); err != nil {
		fmt.Printf("template execute error: %v", err)
		return

	}

}

//testingPage!!! Missleading name, I know, this page takes the info from last page to upload
func testingPage(resp http.ResponseWriter, req *http.Request, encodedMd5 string) {
	/* TODO:
	store file on disk:
	-Accept the file 								DONE
	-create name (from md5)						DONE
	-create a map/index of pub name (hash) to path	(Postponed)
	provide path to file							DONE
	*/
	req.ParseForm()

	upload(resp, req)
}

func upload(resp http.ResponseWriter, req *http.Request) /*(string, error)*/ {
	//encodedMd5 := string()
	//req.ParseForm()
	//fmt.Println("method:", req.Method)
	workingDir, err := os.Getwd()
	keyFile := workingDir + "/keys"
	content, err := ioutil.ReadFile(keyFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	keySplit := strings.Split(string(content), ",")
	if string(content) == "" {
		errorHandler(resp, req, 404)
		return
	}
	if len(keySplit) == 0 {
		log.Fatal("NO KEYS")
		return
	} else {

		for i := 0; i <= len(keySplit)-1; i++ {
			var keySuccess bool
			if req.FormValue("fn") == keySplit[i] {
				keySuccess := true
				if keySuccess != true {
					fmt.Println("Invalid/no key")
					//fmt.Fprintln("NO/INVALID KEY")
					return
				}
				fmt.Printf("Key success!\n")
				if req.Method == "GET" {
					fmt.Println("Yo, its GET for the upload, btw")
					crutime := time.Now().Unix()
					fmt.Println("Beep Beep Beep... The time is:", crutime)
					md5 := md5.New()
					io.WriteString(md5, strconv.FormatInt(crutime, 10))
					bytemd5 := []byte("md5")
					encodedMd5 := hex.EncodeToString(bytemd5)
					fmt.Println("I just got the hashed md5! Here it is:", encodedMd5, "\nEnd of md5sum")
					//fmt.Printf("MD5:", md5)
					token := fmt.Sprintf("%x", md5.Sum(nil))
					t, _ := template.ParseFiles("upload.gtpl")
					t.Execute(resp, token)
					fileURL := "http://" + baseURL + urlPrefix + "i/" + encodedMd5
					http.Redirect(resp, req, fileURL, 301)
					return
				} else {
					req.ParseForm()
					//img := req.FormFile("img")
					fmt.Println("Yo, its POST for the upload, btw")
					req.ParseMultipartForm(32 << 20)
					crutime := time.Now().Unix()
					fmt.Println("Beep Beep Beep... The time is:", crutime)
					file, handler, err := req.FormFile("uploadfile")
					if err != nil {
						fmt.Println(err)
						return
					}
					os.Open(handler.Filename)
					defer file.Close()
					md5 := md5.New()
					io.WriteString(md5, strconv.FormatInt(crutime, 10))
					byteMd5 := []byte(handler.Filename)
					encodedMd5 := hex.EncodeToString(byteMd5)[:imgHash]
					fmt.Println("I just hashed md5! Here it is:", encodedMd5, "\nEnd of md5sum")
					//fmt.Printf("MD5:", md5)ot the
					//token := fmt.Sprintf("%x", md5.Sum(nil))
					//t, _ := template.ParseFiles("upload.gtpl")
					//t.Execute(resp, token)
					firstChar := string(encodedMd5[0])
					secondChar := string(encodedMd5[1])
					//fmt.Printf("File:", file, "\nhandler: ", handler) //too spammy for normal use
					defer file.Close()
					//fmt.Fprintf(resp, "%v", handler.Header)
					os.Open(handler.Filename)
					fmt.Printf("FileName: %s \n", handler.Filename)
					nameSplit := strings.Split(handler.Filename, ".")
					fmt.Printf("File extension: %s\n", nameSplit[len(nameSplit)-1])
					fileName := encodedMd5 + "." + nameSplit[len(nameSplit)-1]
					filepath := path.Join(imgStore, firstChar, secondChar, fileName)
					fmt.Println("file: ", file)

					f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
					fmt.Println("filename?: ", filepath)
					if err != nil {
						fmt.Println(err)
						return
					}
					defer f.Close()
					io.Copy(f, file)
					fmt.Println("Saved file!")
					//sendImg(resp, req, encodedMd5)
					//return encodedMd5, err
					fileURL := baseURL + urlPrefix + "i/" + fileName
					http.Redirect(resp, req, fileURL, http.StatusSeeOther)
					return
				}
			}
			if keySuccess != true {
				fmt.Println("Invalid/no key")
				//fmt.Fprintln("NO/INVALID KEY")
				return
			}

		}

	}

	//return encodedMd5, err
}

// Page for sending pics
func sendImg(resp http.ResponseWriter, req *http.Request, img string) {
	fmt.Println("Recieved a req to send the user a file")
	nameSplit := strings.Split(img, ".")
	imgTitle := nameSplit[0]
	if len(imgTitle) != imgHash {
		//img = defaultImg //if no image exists, use testing image
		//fmt.Println("Using Default Image")
		errorHandler(resp, req, 404)
		fmt.Println("Well this is awkward, our hash is bad, sending 404")
		return
	}
	firstChar := string(img[0])
	secondChar := string(img[1])
	filepath := path.Join(imgStore, firstChar, secondChar, img)
	//Check if file exists and open
	openfile, err := os.Open(filepath)
	defer openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		errorHandler(resp, req, http.StatusNotFound)
		fmt.Printf("ERROR: %s", err)
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
	fmt.Printf("Heres the file size: %s", fileSize)

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
	fmt.Println("artifical http error: ", status)
	fmt.Fprint(resp, "custom ", status)
}

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	fmt.Println("the req arrived")
	if req.Body == nil {
		return
	}
	crutime := time.Now().Unix()

	urlSplit := strings.Split(req.URL.Path, "/")
	urlECount := len(urlSplit)
	fmt.Printf("urlECount: %d\n", urlECount)
	// Checking amt of elements in url (else sends 404)

	if urlECount < 2 {
		errorHandler(resp, req, http.StatusNotFound)
		return
	}
	// Check for prefix
	if !strings.HasPrefix(req.URL.Path, urlPrefix) {
		errorHandler(resp, req, http.StatusNotFound)
		return
	}
	// Defining variables as if there was NO prefix (this will never happen)
	// INSERT FUCTION TO DO VARS

	/* URL Template
	0: usually null, or maybe its the url...?
	1: The prefix, if there is one, if it isnt then this list isnt applicable
	2: The part to ask for a part of the site
	3: data for the site (like file name)
	*/
	// Add your variables involving the URL here
	numberOfPrefixSlashes := strings.Count(urlPrefix, "/") - 1
	switchLen := 1 + numberOfPrefixSlashes
	test1 := 2 + numberOfPrefixSlashes
	test2 := 1 + numberOfPrefixSlashes
	i1 := 2 + numberOfPrefixSlashes
	fmt.Println("The 'info' part of the url is ", switchLen, "\nThe URL is ", req.URL.Path)

	switch urlSplit[switchLen] {
	case "test":
		if urlECount != test1 || urlSplit[test2] == "" {
			errorHandler(resp, req, http.StatusNotFound)
			return
		}
		req.ParseMultipartForm(32 << 20)
		md5 := md5.New()
		fmt.Println("Im in the test case!")
		io.WriteString(md5, strconv.FormatInt(crutime, 10))
		bytemd5 := []byte("md5")
		encodedMd5 := hex.EncodeToString(bytemd5)
		fmt.Println("I just hashed md5! Here it is:", encodedMd5, "\nEnd of md5sum")
		sendImg(resp, req, encodedMd5)
		fmt.Printf("URL is: %v\n", req.URL.Path)
		testingPage(resp, req, encodedMd5)
	case "i":
		// Checks for hash/element/thing
		if urlECount != 4 {
			errorHandler(resp, req, http.StatusNotFound)
			return
		}
		fmt.Printf("urlECount of IMG: %d\n", urlECount)
		fmt.Printf("Split for image: %v\n", urlSplit)
		sendImg(resp, req, urlSplit[3])
		//upload(resp, req)
	case "upload":
		if urlECount != i1 {
			errorHandler(resp, req, http.StatusNotFound)
			return
		}
		fmt.Println("Upload selected")
		upload(resp, req)
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
	fmt.Printf("Heres the prefix for the url, dummy: %s \n", urlPrefix)
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
