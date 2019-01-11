package sessions

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gidoBOSSftw5731/log"
)

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

const (
	allowedChars = "!#$%&'()*+,-./23456789:;<=>?@ABCDEFGHJKLMNOPRSTUVWXYZ[]^_abcdefghijkmnopqrstuvwxyz{|}~" // 86 chars
)

//New is a function to create a new session cookie and write it to the client.
//Im relying on an external system to not overwrite the cookie, though a check will be present.
func New(resp http.ResponseWriter, req *http.Request, sqlPass string) error {
	log.Traceln("Beginning to make a new session for the client")
	lastcookie, _ := req.Cookie("session")
	if lastcookie != nil {
		return fmt.Errorf("SESSION_EXISTS")
	}
	expiration := time.Now().Add(168 * time.Hour)
	allowedCharsSplit := strings.Split(allowedChars, "")
	var session string
	var x int
	rand.Seed(time.Now().UnixNano())
	for i := 0; i <= 128; i++ {
		x = rand.Intn(len(allowedChars)-0-1) + 0 // Not helpful name, but this generates a randon number from 0 to 85 to locate what we need for the session
		session += allowedCharsSplit[x]          // Using x to navigate the split for one character
	}
	cookie := http.Cookie{Name: "session", Value: session, Expires: expiration}

	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ImgSrvr", sqlPass))
	if err != nil {
		log.Error("Oh noez, could not connect to database")
		return fmt.Errorf("Error in SQL! %v", err)
	}
	log.Debug("Oi, mysql did thing")
	defer db.Close()
	var token string
	err = db.QueryRow("SELECT filename FROM sessions WHERE token=?", session).Scan(&token)
	switch {
	case err == sql.ErrNoRows:
		log.Debug("New session, adding..")
		_, err := db.Exec("INSERT INTO sessions VALUES(?, ?, ?, ?)", session, expiration, req.RemoteAddr)
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
