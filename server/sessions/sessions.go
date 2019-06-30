package sessions

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gidoBOSSftw5731/log"
)

// Cookie is a struct detailing all parts of an http cookie
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

func getClientIP(req *http.Request) string {
	ip := req.RemoteAddr
	ipSplit := strings.Split(":", ip)
	return ipSplit[0]
}

const (
	allowedChars = "!#$%&'()*+,-./23456789:<=>?@ABCDEFGHJKLMNOPRSTUVWXYZ[]^_abcdefghijkmnopqrstuvwxyz{|}~" // 85 chars
)

func startSQL(sqlPass string) *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ImgSrvr", sqlPass))
	if err != nil {
		log.Error("Oh noez, could not connect to database")
		log.Errorf("Error in SQL! %v", err)
	}
	log.Debug("Oi, mysql did thing")
	//defer db.Close()

	return db
}

func deleteKey(resp http.ResponseWriter, db *sql.DB, token string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE token = ?;`, token)
	cookie := http.Cookie{Name: "session", Value: "", Expires: time.Now()}
	http.SetCookie(resp, &cookie)
	return err
}

//New is a function to create a new session cookie and write it to the client.
//Im relying on an external system to not overwrite the cookie, though a check
//will be present, returning err SESSION_EXISTS
func New(resp http.ResponseWriter, req *http.Request, sqlPass string) error {
	log.Traceln("Beginning to make a new session for the client")
	lastcookie, _ := req.Cookie("session")
	if lastcookie != nil {
		return fmt.Errorf("SESSION_EXISTS")
	}
	expiration := time.Now().Add(24 * time.Hour).Unix()
	allowedCharsSplit := strings.Split(allowedChars, "")
	var session string
	var x int
	rand.Seed(time.Now().UnixNano())
	for i := 0; i <= 128; i++ {
		x = rand.Intn(len(allowedChars)-0-1) + 0 // Not helpful name, but this generates a randon number from 0 to 84 to locate what we need for the session
		session += allowedCharsSplit[x]          // Using x to navigate the split for one character
	}

	cookie := http.Cookie{Name: "session", Value: session, Expires: time.Unix(expiration, 0)}

	db := startSQL(sqlPass)
	defer db.Close()
	var token string
	err := db.QueryRow("SELECT * FROM sessions WHERE token=?", session).Scan(&token)
	switch {
	case err == sql.ErrNoRows:
		log.Debug("New session, adding..")
		_, err := db.Exec("INSERT INTO sessions VALUES(?, ?, ?)", session, expiration, getClientIP(req))
		if err != nil {
			log.Error(err)
			return err
		}
		log.Debug("Added token info to table")
	case err != nil:
		log.Error(err)
		return err
	default:
		return fmt.Errorf("SQL_ROW_EXISTS")
	}

	http.SetCookie(resp, &cookie)

	return nil
}

//Verify cookies to make sure they aren't expired or invalid.
func Verify(resp http.ResponseWriter, req *http.Request, sqlPass string) (bool, error) {
	log.Traceln("Beginning to check the key")
	OK := true
	cookie, _ := req.Cookie("session")
	if cookie == nil {
		return false, fmt.Errorf("INVALID")
	}

	db := startSQL(sqlPass)
	defer db.Close()

	var expr, ip string
	err := db.QueryRow("SELECT expiration, ip FROM sessions WHERE token=?", cookie.Value).Scan(&expr, &ip)
	switch {
	case err != nil:
		log.Errorln("File not in db..")
		return false, fmt.Errorf("INVALID")
	default:
		log.Traceln("Found a key")
	}
	/*
		if ip != getClientIP(req) {
			OK = false
			err = fmt.Errorf("MISMATCHED_IP")
			return OK, err

		} */

	fmtExpr, _ := strconv.ParseInt(expr, 10, 64)

	if fmtExpr <= time.Now().Unix() {
		err := deleteKey(resp, db, fmt.Sprintln(cookie))
		if err != nil {
			log.Errorln(err)
		}
		return false, fmt.Errorf("EXPIRED")
	}

	return OK, err
}
