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
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/haisum/recaptcha"
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

type files struct {
	Hash        string
	UploaderKey string
	Filename    string
	UploaderIP  string
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

/* TODO:
store file on disk:
-Accept the file 								DONE
-create name (from md5)							DONE
-create a database of pub name (hash) to path	DONE
provide path to file							DONE
*/

func checkKey(resp http.ResponseWriter, req *http.Request, inputKey string) bool {
	workingDir, err := os.Getwd()
	keyFile := workingDir + "/keys"
	content, err := ioutil.ReadFile(keyFile)
	if err != nil {
		fmt.Println(err)
		return false
	}
	keySplit := strings.Split(string(content), ",")
	if string(content) == "" {
		errorHandler(resp, req, 404)
		return false
	}
	if len(keySplit) == 0 {
		log.Fatal("NO KEYS")
	}
	sort.Strings(keySplit)
	n := sort.SearchStrings(keySplit, inputKey)
	if n < len(keySplit) && keySplit[n] == inputKey {
		return true
	}
	return false // last call if all else fails
}

func upload(resp http.ResponseWriter, req *http.Request) /*(string, error)*/ {
	inputKey := req.FormValue("fn")
	//fmt.Println("[DEBUG ONLY] Key is:", inputKey) // have this off unless testing
	re := recaptcha.R{
		Secret: recaptchaPrivKey,
	}
	isValid := re.Verify(*req)
	if !isValid {
		fmt.Fprintf(resp, "Invalid! These errors ocurred: %v", re.LastError())
		fmt.Printf("Invalid! These errors ocurred: %v", re.LastError())
		return
	}

	if checkKey(resp, req, inputKey) == true {
		fmt.Printf("Key success!\n")

	} else {
		fmt.Println("Invalid/no key")
		//fmt.Fprintln("NO/INVALID KEY")
		return
	}
	if req.Method == "GET" {
		/*fmt.Println("Yo, its GET for the upload, btw")
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
		fileURL := baseURL + urlPrefix + "i/" + encodedMd5
		http.Redirect(resp, req, fileURL, 301)
		return*/
		fmt.Fprintln(resp, "GET IS NOT SUPPORTED") /* Im too lazy to add GET support
		and it will never occur, its just a dying branch of code*/
		return
	} else {
		db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ImgSrvr", sqlPasswd))
		if err != nil {
			fmt.Println("Oh noez, could not connect to database")
			return
		}
		fmt.Println("Oi, mysql did thing")
		defer db.Close()
		if err != nil {
			fmt.Println("Oh noez, could not connect to database")
			return
		} // end of SQL opening
		req.ParseMultipartForm(32 << 20)
		req.ParseForm()
		//img := req.FormFile("img")
		fmt.Println("Yo, its POST for the upload, btw")
		crutime := time.Now().Unix()
		fmt.Println("Beep Beep Beep... The time is:", crutime)
		file, handler, err := req.FormFile("uploadfile") // Saving file to memory
		defer file.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("file: ", file) //Although this means nothing itself, its nice to have in case its a 0 byte file

		md5 := md5.New() //Make a MD5 variable, to be changed later... maybe..
		written, err := io.Copy(md5, file)
		if err != nil {
			fmt.Println(err)
			return
		}
		if written == 0 {
			fmt.Println("No file written, error!: ", written)
			return
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		encodedMd5 := hex.EncodeToString(md5.Sum(nil))[:imgHash]
		fmt.Println("I just hashed md5! Here it is:", encodedMd5)
		firstChar := string(encodedMd5[0])
		secondChar := string(encodedMd5[1])
		fmt.Println("FileName: \n", handler.Filename)
		var sqlFilename string
		err = db.QueryRow("SELECT filename FROM files WHERE hash=?", encodedMd5).Scan(&sqlFilename)
		switch {
		case err == sql.ErrNoRows:
			fmt.Println("New file, adding..")
			insert, err := db.Query("INSERT INTO files VALUES(?, ?, ?, ?)", encodedMd5, inputKey, handler.Filename, req.RemoteAddr)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer insert.Close()
			fmt.Println("Added fiel info to table")
			sqlFilename = handler.Filename
		case err != nil:
			fmt.Println(err)
			return
		default:
		}
		filepath := path.Join(imgStore, firstChar, secondChar, sqlFilename)
		f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666) // WRONLY means Write only
		fmt.Println("filename?: ", filepath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		written, err = io.Copy(f, file)
		if err != nil {
			fmt.Println(err)
			return
		}
		if written == 0 {
			fmt.Println("No file written, error!: ", written)
			return
		}
		fmt.Println("Saved file!")
		fileURL := baseURL + urlPrefix + "i/" + encodedMd5
		http.Redirect(resp, req, fileURL, http.StatusSeeOther)
		return
	}

	//return encodedMd5, err
}

// Page for sending pics
func sendImg(resp http.ResponseWriter, req *http.Request, img string) {
	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ImgSrvr", sqlPasswd))
	if err != nil {
		fmt.Println("Oh noez, could not connect to database")
		return
	}
	fmt.Println("Oi, mysql did thing")
	defer db.Close()
	if err != nil {
		fmt.Println("Oh noez, could not connect to database")
		return
	} // end of SQL opening
	fmt.Println("Recieved a req to send the user a file")
	if len(img) != imgHash {
		//img = defaultImg //if no image exists, use testing image
		//fmt.Println("Using Default Image")
		errorHandler(resp, req, 404)
		fmt.Println("Well this is awkward, our hash is bad, sending 404")
		return
	}
	firstChar := string(img[0])
	secondChar := string(img[1])

	if err != nil {
		//File not found, send 404
		errorHandler(resp, req, http.StatusNotFound)
		fmt.Printf("ERROR: %s", err)
		return
	}
	var filename string
	err = db.QueryRow("SELECT filename FROM files WHERE hash=?", img).Scan(&filename)
	switch {
	case err == sql.ErrNoRows:
		fmt.Println("File not in db..")
	case err != nil:
		fmt.Println(err)
		return
	default:
		fmt.Println("Filename from sql is:", filename)
	}
	filepath := path.Join(imgStore, firstChar, secondChar, filename)
	//Check if file exists and open
	openfile, err := os.Open(filepath)
	defer openfile.Close() //Close after function return
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
	ip := req.RemoteAddr
	fmt.Println("This request is being requested by:", ip)

	urlSplit := strings.Split(req.URL.Path, "/")
	urlECount := len(urlSplit)
	fmt.Println("The url is:", req.URL.Path)
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
	//test1 := 2 + numberOfPrefixSlashes
	//test2 := 1 + numberOfPrefixSlashes
	i1 := 2 + numberOfPrefixSlashes

	switch urlSplit[switchLen] {
	/*case "test":
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
	testingPage(resp, req, encodedMd5)*/
	case "i":
		// Checks for hash/element/thing
		if urlSplit[i1] == "" {
			errorHandler(resp, req, http.StatusNotFound)
			return
		}
		fmt.Printf("urlECount of IMG: %d\n", urlECount)
		fmt.Printf("Split for image: %v\n", urlSplit)
		sendImg(resp, req, urlSplit[i1])
		//upload(resp, req)
	case "upload":
		fmt.Println("Upload selected")
		upload(resp, req)
	case "favicon.ico", "favicon-16x16.png", "favicon-32x32.png", "favicon-96x96.png", "favicon-256x256" +
		".png", "android-icon-192x192.png", "apple-icon-114x114.png", "apple-icon-120x120.png", "apple-icon-" +
		"144x144.png", "apple-icon-152x152.png", "apple-icon-180x180.png", "apple-icon-57x57.png", "apple-icon-" +
		"60x60.png", "apple-icon-72x72.png", "apple-icon-76x76.png", "ms-icon-144x144.png", "ms-icon-150x150" +
		".png", "ms-icon-310x310.png", "ms-icon-70x70.png": // case for favicons
		http.ServeFile(resp, req, "favicons/"+urlSplit[switchLen])
	case "robots.txt":
		http.ServeFile(resp, req, "robots.txt")
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
	// I reccomend blocking 3306 in your firewall unless you use the port elsewhere
	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ImgSrvr", sqlPasswd))
	if err != nil {
		fmt.Println("Oh noez, could not connect to database")
		return
	}

	fmt.Println("Oi, mysql did thing")
	defer db.Close()

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
