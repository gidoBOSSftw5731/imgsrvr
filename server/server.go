package server

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	//raven "github.com/getsentry/raven-go"
	"github.com/gidoBOSSftw5731/log"
	"github.com/haisum/recaptcha"
)

//tData is a struct for HTTP inputs.
type tData struct {
	Fn string
	Tn string
}

//FileHeader is used for when you download a file from the client. It stores all relevant information in Header.
type FileHeader struct {
	Filename string
	Header   textproto.MIMEHeader
	// contains filtered or unexported fields
}

//files is used for implementing user-data into SQL databases... in theory..
type files struct {
	Hash        string
	UploaderKey string
	Filename    string
	UploaderIP  string
}

//Cookie is a struct for creating data for cookies
type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}

//Config is a struct for importing the config from main.go
type config struct {
	urlPrefix, imgStore, baseURL, sqlPasswd, recaptchaPrivKey, recaptchaPubKey, coinhiveCaptchaPub, coinhiveCaptchaPriv string
	imgHash                                                                                                             int
}

//FastCGIServer is how the config constants get to the server package.
type FastCGIServer struct {
	config config
}

//NewFastCGIServer is an implementation of fastcgi server.
func NewFastCGIServer(urlPrefix, imgStore, baseURL, sqlPasswd, recaptchaPrivKey, recaptchaPubKey, coinhiveCaptchaPub, coinhiveCaptchaPriv string, imgHash int) *FastCGIServer {
	return &FastCGIServer{
		config: config{
			urlPrefix:           urlPrefix,
			imgHash:             imgHash,
			imgStore:            imgStore,
			baseURL:             baseURL,
			sqlPasswd:           sqlPasswd,
			recaptchaPrivKey:    recaptchaPrivKey,
			recaptchaPubKey:     recaptchaPubKey,
			coinhiveCaptchaPub:  coinhiveCaptchaPub,
			coinhiveCaptchaPriv: coinhiveCaptchaPriv,
		}}
}

/*HTML meaning guide for my sanity:
<br>: page break
<html> and </html>: beginning and end of html section
<body> and </body>: main part with text
<!-- and -->: Comments... WHY
<a href="url"> and </a>: links
<p> and </p>: margin
*/

//cookieCheck is a func implemented in every Page func to handle the cookies.
func cookieCheck(resp http.ResponseWriter, req *http.Request, config config) {
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{Name: "ip", Value: req.RemoteAddr, Expires: expiration}
	prevCookie, _ := req.Cookie("ip")
	log.Traceln("Last IP was: ", prevCookie)
	http.SetCookie(resp, &cookie)
}

/*func prepareTemplate(source string) (*template.New, string, error) {

}*/

//todoPage is a standard func for the setup of the todo page.
func todoPage(resp http.ResponseWriter, req *http.Request, config config) {
	cookieCheck(resp, req, config)
	todoPageTemplate := template.New("first page templated.")
	content, err := ioutil.ReadFile("server/todoPageVar.html")
	todoPageVar := string(content)
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		errorHandler(resp, req, 404)
		return
	}
	todoPageTemplate, err = todoPageTemplate.Parse(fmt.Sprintf(todoPageVar, config.urlPrefix, config.urlPrefix, config.urlPrefix))
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		errorHandler(resp, req, 404)
		return
	}
	field := req.FormValue("tn")
	tData := tData{
		Fn: field,
	}
	err = todoPageTemplate.Execute(resp, tData)
}

//appPage is a standard func for the setup of the main page.
func appPage(resp http.ResponseWriter, req *http.Request, config config) {
	cookieCheck(resp, req, config)
	firstPageTemplate := template.New("first page templated.")
	content, err := ioutil.ReadFile("server/firstPage.html")
	firstPage := string(content)
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		errorHandler(resp, req, 404)
		return
	}
	firstPageTemplate, err = firstPageTemplate.Parse(fmt.Sprintf(firstPage, config.urlPrefix, config.recaptchaPubKey,
		config.coinhiveCaptchaPub, config.urlPrefix, config.urlPrefix, config.urlPrefix))
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		return
	}
	req.ParseForm()
	field := req.FormValue("fn")
	//fmt.Println(field)
	tData := tData{
		Fn: field,
	}
	//upload(resp, req)
	//log.Traceln("Form data: ", field, "\ntData: ", tData)
	err = firstPageTemplate.Execute(resp, tData)
	if err != nil {
		log.Errorf("template execute error: %v", err)
		return

	}

}

//megaMinePage is a standard func for the setup of the megamine page.
func megaMinePage(resp http.ResponseWriter, req *http.Request, config config) {
	megaMineTemplate := template.New("first page templated.")
	content, err := ioutil.ReadFile("server/megaMineVar.html")
	megaMineVar := string(content)
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		errorHandler(resp, req, 404)
		return
	}
	megaMineTemplate, err = megaMineTemplate.Parse(fmt.Sprintf(megaMineVar, config.urlPrefix, config.urlPrefix,
		config.urlPrefix, config.urlPrefix))
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		return
	}
	field := req.FormValue("mn")
	tData := tData{
		Fn: field,
	}
	err = megaMineTemplate.Execute(resp, tData)
}

func miningPage(resp http.ResponseWriter, req *http.Request, config config) {
	cookieCheck(resp, req, config)
	minePageTemplate := template.New("first page templated.")
	content, err := ioutil.ReadFile("server/minePageVar.html")
	minePageVar := string(content)
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		errorHandler(resp, req, 404)
		return
	}
	minePageTemplate, err = minePageTemplate.Parse(fmt.Sprintf(minePageVar, config.urlPrefix, config.urlPrefix,
		config.urlPrefix, config.urlPrefix))
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		return
	}
	field := req.FormValue("mn")
	tData := tData{
		Fn: field,
	}
	err = minePageTemplate.Execute(resp, tData)
}

//checkKey is a function to read the key given by the user and check it against a list of known-good keys
//to ensure validity of the key before handling any sensitive parts of the system.
//needs optimization
func checkKey(resp http.ResponseWriter, req *http.Request, inputKey string) bool {
	workingDir, err := os.Getwd()
	keyFile := workingDir + "/keys"
	content, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Errorln(err)
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

//verifyCCaptcha is a func to check the validity of the response from the coinhive captcha
func verifyCCaptcha(resp http.ResponseWriter, req *http.Request, config config) (string, error) {
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

	w, err := http.PostForm("https://api.coinhive.com/token/verify",
		url.Values{"secret": {config.coinhiveCaptchaPriv}, "hashes": {"1024"}, "token": {req.FormValue("coinhive-captcha-token")}})
	if err != nil {
		return "", fmt.Errorf("Error while making captcha request: %v", err)
	}
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	/*w, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", fmt.Errorf("Error while making captcha request: %v", err)
	}*/
	defer w.Body.Close()

	wBody, err := ioutil.ReadAll(w.Body)
	if err != nil {
		return "", fmt.Errorf("Error while making captcha request: %v", err)
	}
	//log.Traceln(string(wBody))
	//log.Traceln("response Status:", w.Status)
	//log.Traceln("response Headers:", w.Header)
	//log.Traceln("response Body:", string(body))
	return string(wBody), nil
}

//upload is the func to take the users file  and upload it.
func upload(resp http.ResponseWriter, req *http.Request, config config) /*(string, error)*/ {
	inputKey := req.FormValue("fn")

	fmt.Println("got into the func")

	//fmt.Println("[DEBUG ONLY] Key is:", inputKey) // have this off unless testing
	re := recaptcha.R{
		Secret: config.recaptchaPrivKey,
	}
	coinhiveResp, err := verifyCCaptcha(resp, req, config)
	if err != nil {
		//File not found, send 404
		errorHandler(resp, req, 404)
		log.Errorf("ERROR: %s", err)
		return
	}
	log.Tracef("Coinhive's response: %v\n", coinhiveResp)
	isValid2 := strings.HasPrefix(coinhiveResp, `{"success":true,"created":`) // coinhive
	isValid := re.Verify(*req)                                                // recaptcha
	if !isValid && !isValid2 {
		fmt.Fprintf(resp, "Invalid Captcha! These errors ocurred: %v", re.LastError())
		fmt.Printf("Invalid Captcha! These errors ocurred: %v", re.LastError())
		return
	} else {
		log.Traceln("recieved a valid captcha response!")
	}

	if checkKey(resp, req, inputKey) == true {
		log.Debugln("Key success!\n")

	} else {
		log.Errorln("Invalid/no key")
		fmt.Fprintln(resp, "Invalid/No key!!!")
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
		db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ImgSrvr", config.sqlPasswd))
		if err != nil {
			log.Error("Oh noez, could not connect to database")
			return
		}
		log.Debug("Oi, mysql did thing")
		defer db.Close()
		// end of SQL opening
		err = req.ParseMultipartForm(107374182400) // max upload in... bytes..?
		if err != nil {
			errorHandler(resp, req, http.StatusBadRequest)
			log.Errorf("File too Big! err = %v", err)
			return
		}
		req.ParseForm()
		//img := req.FormFile("img")
		log.Trace("Yo, its POST for the upload, btw")
		crutime := time.Now().Unix()
		log.Trace("Beep Beep Beep... The time is:", crutime)
		file, handler, err := req.FormFile("uploadfile") // Saving file to memory
		switch err {
		case nil:
		case http.ErrMissingFile:
			log.Error("NO FILE")
			fmt.Fprintln(resp, "NO FILE")
			return
		default:
			log.Error(err)
			errorHandler(resp, req, 404)
			return
		}
		defer file.Close()
		log.Trace("file: ", file) //Although this means nothing itself, its nice to have in case its a 0 byte file

		md5 := md5.New() //Make a MD5 variable, to be changed later... maybe..
		written, err := io.Copy(md5, file)
		if err != nil {
			log.Error(err)
			return
		}
		if written == 0 {
			log.Error("No md5 written, error!: ", written)
			return
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			log.Error(err)
			return
		}
		encodedMd5 := hex.EncodeToString(md5.Sum(nil))[:config.imgHash]
		log.Trace("I just hashed md5! Here it is:", encodedMd5)
		firstChar := string(encodedMd5[0])
		secondChar := string(encodedMd5[1])
		log.Tracef("FileName: %v\n", handler.Filename)
		var sqlFilename string
		err = db.QueryRow("SELECT filename FROM files WHERE hash=?", encodedMd5).Scan(&sqlFilename)
		switch {
		case err == sql.ErrNoRows:
			log.Debug("New file, adding..")
			_, err := db.Exec("INSERT INTO files VALUES(?, ?, ?, ?)", encodedMd5, inputKey, handler.Filename, req.RemoteAddr) // the _ var used to be `insert` but was removed due to an issue
			if err != nil {
				log.Error(err)
				return
			}
			//defer insert.Close()
			log.Debug("Added fiel info to table")
			sqlFilename = handler.Filename
		case err != nil:
			log.Error(err)
			return
		default:
		}
		filepath := path.Join(config.imgStore, firstChar, secondChar, sqlFilename)
		f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666) // WRONLY means Write only
		log.Trace("filename?: ", filepath)
		if err != nil {
			log.Error(err)
			return
		}
		defer f.Close()
		written, err = io.Copy(f, file)
		if err != nil {
			log.Error(err)
			return
		}
		if written == 0 {
			log.Error("No file written, error!: ", written)
			return
		}
		log.Infof("Saved file at %v!", crutime)
		fileURL := config.baseURL + config.urlPrefix + "i/" + encodedMd5
		http.Redirect(resp, req, fileURL, http.StatusSeeOther)
		return
	}

	//return encodedMd5, err
}

// Page for sending pics
func sendImg(resp http.ResponseWriter, req *http.Request, img string, config config) {
	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ImgSrvr", config.sqlPasswd))
	if err != nil {
		log.Errorln("Oh noez, could not connect to database")
		return
	}
	log.Traceln("Oi, mysql did thing")

	if err != nil {
		log.Errorln("Oh noez, could not connect to database")
		return
	}
	defer db.Close() // end of SQL opening
	imgSplit := strings.Split(img, ".")
	img = imgSplit[0]
	log.Traceln("Recieved a req to send the user a file")
	if len(img) != config.imgHash {
		//img = defaultImg //if no image exists, use testing image
		//fmt.Println("Using Default Image")
		errorHandler(resp, req, 404)
		log.Errorln("Well this is awkward, our hash is bad, sending 404")
		return
	}
	firstChar := string(img[0])
	secondChar := string(img[1])

	if err != nil {
		//File not found, send 404
		errorHandler(resp, req, http.StatusNotFound)
		log.Errorf("ERROR: %s", err)
		return
	}
	var filename string
	err = db.QueryRow("SELECT filename FROM files WHERE hash=?", img).Scan(&filename)
	switch {
	case err == sql.ErrNoRows:
		log.Errorln("File not in db..")
		errorHandler(resp, req, 404)
		return
	case err != nil:
		log.Errorln(err)
		errorHandler(resp, req, 404)
		return
	default:
		log.Traceln("Filename from sql is:", filename)
	}
	filepath := path.Join(config.imgStore, firstChar, secondChar, filename)
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
	log.Tracef("Heres the file size: %s", fileSize)

	//Send the headers
	resp.Header().Set("Content-Disposition", "attachment; filename="+filename)
	resp.Header().Set("Content-Type", fileContentType)
	resp.Header().Set("Content-Length", fileSize)

	//Send the file
	//We read 512 bytes from the file already so we reset the offset back to 0
	openfile.Seek(0, 0)
	written, err := io.Copy(resp, openfile) //'Copy' the file to the client
	if err != nil {
		log.Error(err)
		return
	}
	if written == 0 {
		log.Error("No file written, error!: ", written)
		return
	}
	log.Traceln("Successful upload")
	return
}

func errorHandler(resp http.ResponseWriter, req *http.Request, status int) {
	resp.WriteHeader(status)
	log.Error("artifical http error: ", status)
	fmt.Fprint(resp, "custom ", status)
}

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	log.Debug("the req arrived")
	if req.Body == nil {
		return
	}
	ip := req.RemoteAddr
	log.Debug("This request is being requested by:", ip)

	urlSplit := strings.Split(req.URL.Path, "/")
	urlECount := len(urlSplit)
	log.Debug("The url is:", req.URL.Path)
	log.Debugf("urlECount: %d\n", urlECount)
	// Checking amt of elements in url (else sends 404)
	//Check for prefix
	if urlECount < 2 || !strings.HasPrefix(req.URL.Path, s.config.urlPrefix) {
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
	numberOfPrefixSlashes := strings.Count(s.config.urlPrefix, "/") - 1
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
		log.Tracef("urlECount of IMG: %d\n", urlECount)
		log.Tracef("Split for image: %v\n", urlSplit)
		sendImg(resp, req, urlSplit[i1], s.config)
		//upload(resp, req)
	case "miner":
		miningPage(resp, req, s.config)
	case "megamine":
		megaMinePage(resp, req, s.config)
	case "upload":
		log.Traceln("Upload selected")
		upload(resp, req, s.config)
	case "todo":
		todoPage(resp, req, s.config)
	case "favicon.ico", "favicon-16x16.png", "favicon-32x32.png", "favicon-96x96.png", "favicon-256x256" +
		".png", "android-icon-192x192.png", "apple-icon-114x114.png", "apple-icon-120x120.png", "apple-icon-" +
		"144x144.png", "apple-icon-152x152.png", "apple-icon-180x180.png", "apple-icon-57x57.png", "apple-icon-" +
		"60x60.png", "apple-icon-72x72.png", "apple-icon-76x76.png", "ms-icon-144x144.png", "ms-icon-150x150" +
		".png", "ms-icon-310x310.png", "ms-icon-70x70.png": // case for favicons
		http.ServeFile(resp, req, "favicons/"+urlSplit[switchLen])
	case "robots.txt":
		http.ServeFile(resp, req, "robots.txt")
	case "css":
		http.ServeFile(resp, req, "server/"+urlSplit[switchLen+1])
	case "js":
		i := switchLen + 2
		if i >= len(urlSplit)+1 {
			errorHandler(resp, req, 404)
			return
		}
		buf := "/"
		for i <= len(urlSplit) {
			buf += urlSplit[i-1]
			i++
		}
		http.ServeFile(resp, req, path.Join("js/", buf))
	case "":
		//raven.RecoveryHandler(appPage(resp, req, s.config))
		appPage(resp, req, s.config)
	default:
		errorHandler(resp, req, 404)
	}

}
