package tools

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"../../sessions"

	"github.com/gidoBOSSftw5731/log"
	"github.com/haisum/recaptcha"
	"golang.org/x/crypto/bcrypt"
)

var (
	keyFilename = "keys"
	keys        = make(map[string]bool)
)

//hashable is a struct of all the information necessary to check a password hash
type hashable struct {
	key, salt, origHash *string
	pepper              string
}

//hashes is a struct to hold an array of hashes, a few other details are passed for later processing.
type hashes struct {
	arr [52]hashable
	ok  bool
	wg  *sync.WaitGroup
}

//Config is a struct for importing the config from main.go
type Config struct {
	URLPrefix, ImgStore, BaseURL, SQLAcc, RecaptchaPrivKey, RecaptchaPubKey string
	ImgHash                                                                 int
}

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	cost     = 15
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

//AppPage is a standard func for the setup of the main page.
func AppPage(resp http.ResponseWriter, req *http.Request, config Config) {
	firstPageTemplate := template.New("first page templated.")
	content, err := ioutil.ReadFile("server/firstPage.html")
	firstPage := string(content)
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		ErrorHandler(resp, req, 404)
		return
	}
	firstPageTemplate, err = firstPageTemplate.Parse(fmt.Sprintf(firstPage, config.URLPrefix, config.RecaptchaPubKey,
		config.URLPrefix))
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		return
	}
	req.ParseForm()

	tData := tData{ //template data
		Fn: "",
	}
	//upload(resp, req)
	//log.Traceln("Form data: ", field, "\ntData: ", tData)
	err = firstPageTemplate.Execute(resp, tData)
	if err != nil {
		log.Errorf("template execute error: %v", err)
		return

	}

}

//Directory is a function that opens up the directory of things.
func Directory(resp http.ResponseWriter, req *http.Request, config Config) {
	firstPageTemplate := template.New("first page templated.")
	content, err := ioutil.ReadFile("server/selector/modules/directory.html")
	page := string(content)
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		ErrorHandler(resp, req, 404)
		return
	}
	firstPageTemplate, err = firstPageTemplate.Parse(Page)
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		return
	}
	err = firstPageTemplate.Execute(resp, nil)
	if err != nil {
		log.Errorf("template execute error: %v", err)
		return

	}

}

//SignIn is a function to generate a sign in page
func SignIn(resp http.ResponseWriter, req *http.Request, config Config) {
	pageTemplate := template.New("signin page templated.")
	content, err := ioutil.ReadFile("server/signin.html")
	page := string(content)
	if err != nil {
		log.Errorf("Failed to parse template: %v", err)
		ErrorHandler(resp, req, 404)
		return
	}
	pageTemplate, err = pageTemplate.Parse(fmt.Sprintf(page, config.URLPrefix, config.RecaptchaPubKey, config.URLPrefix))
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
	err = pageTemplate.Execute(resp, tData)
	if err != nil {
		log.Errorf("template execute error: %v", err)
		return

	}
}

//LoginHandler handles the login on the login page
func LoginHandler(resp http.ResponseWriter, req *http.Request, config Config) {
	log.Traceln("logging someone in!")

	req.ParseForm()

	captcha, err := checkCaptcha(req, config.RecaptchaPrivKey)
	if err != nil || !captcha {
		ErrorHandler(resp, req, 429)
		log.Errorf("Wrong Captcha = %v", err)
		return
	}

	user := req.FormValue("user")
	_, ok := checkKey(resp, req, req.FormValue("fn"), config.SQLAcc, &user, true)
	if ok {
		http.Redirect(resp, req, config.BaseURL+"/", 302)
	} else {
		http.Redirect(resp, req, config.BaseURL+"/login"+"?issue=BadUserPass", 302)
		return
	}

}

//SendImg is a page for sending pics
func SendImg(resp http.ResponseWriter, req *http.Request, img string, config Config) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s@tcp(127.0.0.1:3306)/ImgSrvr", config.SQLAcc))
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
	if len(img) != config.ImgHash {
		//img = defaultImg //if no image exists, use testing image
		//fmt.Println("Using Default Image")
		ErrorHandler(resp, req, 404)
		log.Errorln("Well this is awkward, our hash is bad, sending 404")
		return
	}
	firstChar := string(img[0])
	secondChar := string(img[1])

	if err != nil {
		//File not found, send 404
		ErrorHandler(resp, req, http.StatusNotFound)
		log.Errorf("ERROR: %s", err)
		return
	}
	var filename string
	err = db.QueryRow("SELECT filename FROM files WHERE hash=?", img).Scan(&filename)

	switch {
	case err == sql.ErrNoRows:
		log.Errorln("File not in db..")
		ErrorHandler(resp, req, 404)
		return
	case err != nil:
		log.Errorln(err)
		ErrorHandler(resp, req, 404)
		return
	default:
		log.Traceln("Filename from sql is:", filename)
	}
	filepath := path.Join(config.ImgStore, firstChar, secondChar, filename)
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
	resp.Header().Set("Content-Disposition", "inline;"+fmt.Sprintf("filename=\"%v\"", filename))
	resp.Header().Set("Content-Type", fileContentType)
	resp.Header().Set("Content-Length", fileSize)
	//resp.AppendHeader("content-disposition", "attachment; filename=\"" + filename +"\"");

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

//ErrorHandler is a function to handle HTTP errors
func ErrorHandler(resp http.ResponseWriter, req *http.Request, status int) {
	resp.WriteHeader(status)
	log.Error("artifical http error: ", status)
	fmt.Fprint(resp, "custom ", status)
}

// ReadKeys reads a key file from disk, returning a map to use in verification.
func ReadKeys(kf string) error {
	// Read the key content from the full file path.
	content, err := ioutil.ReadFile(kf)
	if err != nil {
		log.Errorln(err)
		return err
	}

	// Fill the map with keys seen.
	for _, key := range strings.Split(string(content), ",") {
		keys[key] = true
	}
	return nil
}

func chkHash(inout chan *hashes) {
	var output hashes
	input := <-inout

	var ok bool

	var wg0 sync.WaitGroup
	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup
	var wg3 sync.WaitGroup

	wg0.Add(len(alphabet) / 4)
	wg1.Add(len(alphabet) / 4)
	wg2.Add(len(alphabet) / 4)
	wg3.Add(len(alphabet) / 4)

	go func() {
		for i := 0; i < len(alphabet)/4; i++ {
			obj := input.arr[i]

			if ok {
				wg0.Done()
				continue
			}

			err := bcrypt.CompareHashAndPassword([]byte(*obj.origHash), []byte(string(obj.pepper+*obj.key+*obj.salt)))
			//log.Traceln(string(obj.pepper), string(*obj.origHash))

			if err == nil {
				ok = true
			}
			output.arr[i] = obj

			wg0.Done()
		}
	}()
	go func() {
		for i := len(alphabet) / 4; i < 2*len(alphabet)/4; i++ {
			obj := input.arr[i]

			if ok {
				wg1.Done()
				continue
			}

			err := bcrypt.CompareHashAndPassword([]byte(*obj.origHash), []byte(string(obj.pepper+*obj.key+*obj.salt)))
			//log.Traceln(string(obj.pepper), string(*obj.origHash))

			if err == nil {
				ok = true
			}
			output.arr[i] = obj

			wg1.Done()
		}
	}()
	go func() {
		for i := 2 * len(alphabet) / 4; i < 3*len(alphabet)/4; i++ {
			obj := input.arr[i]

			if ok {
				wg2.Done()
				continue
			}

			err := bcrypt.CompareHashAndPassword([]byte(*obj.origHash), []byte(string(obj.pepper+*obj.key+*obj.salt)))
			//log.Traceln(string(obj.pepper), string(*obj.origHash))

			if err == nil {
				ok = true
			}
			output.arr[i] = obj

			wg2.Done()
		}
	}()
	go func() {
		for i := 3 * len(alphabet) / 4; i < len(alphabet); i++ {
			obj := input.arr[i]

			if ok {
				wg3.Done()
				continue
			}

			err := bcrypt.CompareHashAndPassword([]byte(*obj.origHash), []byte(string(obj.pepper+*obj.key+*obj.salt)))
			//log.Traceln(string(obj.pepper), string(*obj.origHash))

			if err == nil {
				ok = true
			}
			output.arr[i] = obj

			wg3.Done()
		}
	}()

	wg0.Wait()
	wg1.Wait()
	wg2.Wait()
	wg3.Wait()

	output.ok = ok

	log.Debugln("Done checking password!")

	input.wg.Done()

	inout <- &output
}

func checkHash(key, user string, db *sql.DB) (bool, error) {
	var ok bool

	//fmt.Println(user)
	var origHash, salt string
	var in hashes
	err := db.QueryRow("SELECT hash, salt FROM users WHERE user=?", user).Scan(&origHash, &salt)
	if err != nil {
		log.Errorln(err)
		return ok, err
	}

	c := make(chan *hashes)

	var wg sync.WaitGroup

	go chkHash(c)

	in.wg = &wg
	wg.Add(1)

	for i, x := range alphabet {
		in.arr[i] = hashable{&key, &salt, &origHash, string(x)}
	}
	c <- &in

	wg.Wait()

	output := *<-c
	if output.ok {
		ok = true
	}

	log.Debugln("Password Success: ", ok)

	return ok, err
}

// checkKey simply looks in the keys map for evidence of a key.
func checkKey(resp http.ResponseWriter, req *http.Request, inputKey, sqlAcc string, user *string, newSess bool) (bool, bool) { // session good, key good
	ok, err := sessions.Verify(resp, req, sqlAcc, user) // good session
	if ok {
		return true, true
	}
	if err != nil {
		log.Errorln(err)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s@tcp(127.0.0.1:3306)/ImgSrvr", sqlAcc))
	if err != nil {
		log.Errorln("Oh noez, could not connect to database")
		return false, false
	}
	log.Traceln("Oi, mysql did thing")

	keyOK, err := checkHash(inputKey, *user, db)

	if !keyOK { // key not good
		return false, false
	}

	if newSess {
		err = sessions.New(resp, req, sqlAcc) // make new session if none found and valid key
		if err != nil {
			log.Errorln(err)

			switch err.Error() {
			case "SESSION_EXISTS", "":
			default:
				return false, false
			}
		}
	}

	return false, true // session bad key good
}

func checkCaptcha(req *http.Request, priv string) (bool, error) {
	var err error

	re := recaptcha.R{
		Secret: priv,
	}
	//req2 := req
	isValid := re.Verify(*req) // recaptcha
	if !isValid {
		//fmt.Fprintf(resp, "Invalid Captcha! These errors ocurred: %v", re.LastError())
		err = fmt.Errorf("Invalid Captcha! These errors ocurred: %v", re.LastError())
	} else {
		log.Traceln("recieved a valid captcha response!")
	}

	if false { // solely for testing, since I sometimes work offline, should be false on prod machines
		isValid = true
		err = nil
	}
	return isValid, err
}

//Upload is the func to take the users file  and upload it.
func Upload(resp http.ResponseWriter, req *http.Request, config Config) /*(string, error)*/ {
	//fmt.Println("[DEBUG ONLY] Key is:", inputKey) // have this off unless testing

	req.ParseMultipartForm(2 ^ 64 - 1)
	req.Header.Add("Content-Type", "multipart/form-data")

	inputKey := req.FormValue("fn")
	user := req.FormValue("user")

	captcha, err := checkCaptcha(req, config.RecaptchaPrivKey)
	if err != nil || !captcha {
		if err != nil {
			ErrorHandler(resp, req, 429)
			log.Errorf("Wrong Captcha = %v", err)
			return
		}
	}

	//err = req.ParseMultipartForm(107374182400) // max upload in... bytes..?
	/*err = req.ParseForm()
	if err != nil {
		ErrorHandler(resp, req, http.StatusBadRequest)
		log.Errorf("File too Big! err = %v", err)
		return
	}*/

	sessionGood, keyGood := checkKey(resp, req, inputKey, config.SQLAcc, &user, false)

	if sessionGood || keyGood {
		log.Debugln("Key success!")
	} else {
		log.Errorln("Invalid/no key")
		fmt.Fprintln(resp, "Invalid/No key!!!")
		return
	}

	if !sessionGood {

	}
	if req.Method == "POST" {
		db, err := sql.Open("mysql", fmt.Sprintf("%s@tcp(127.0.0.1:3306)/ImgSrvr", config.SQLAcc))
		if err != nil {
			log.Error("Oh noez, could not connect to database")
			return
		}
		log.Debug("Oi, mysql did thing")
		defer db.Close()
		// end of SQL opening

		//req.ParseForm()
		//img := req.FormFile("img")
		log.Trace("It's POST for the upload")
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
			ErrorHandler(resp, req, 404)
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
		encodedMd5 := hex.EncodeToString(md5.Sum(nil))[:config.ImgHash]
		log.Trace("I just hashed md5! Here it is:", encodedMd5)
		firstChar := string(encodedMd5[0])
		secondChar := string(encodedMd5[1])
		log.Tracef("FileName: %v\n", handler.Filename)
		var sqlFilename string
		err = db.QueryRow("SELECT filename FROM files WHERE hash=?", encodedMd5).Scan(&sqlFilename)
		switch {
		case err == sql.ErrNoRows:
			log.Debug("New file, adding..")
			_, err := db.Exec("INSERT INTO files VALUES(?, ?, ?, ?)", encodedMd5, user, handler.Filename, req.RemoteAddr) // the _ var used to be `insert` but was removed due to an issue
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
		filepath := path.Join(config.ImgStore, firstChar, secondChar, sqlFilename)
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
		fileURL := config.BaseURL + config.URLPrefix + "i/" + encodedMd5
		http.Redirect(resp, req, fileURL, http.StatusSeeOther)
	} else {
		fmt.Fprintln(resp, "POST requests only")
	}
	return
	//return encodedMd5, err
}
