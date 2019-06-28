package hasholdkeys

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"../../server"
	"github.com/gidoBOSSftw5731/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	keys = make(map[string]bool)
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	cost     = 20
)

func randInt() int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(52))
	if err != nil {
		panic(err)
	}
	n := nBig.Int64()
	return n
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

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

// Run is a function to hash legacy keys.
func Run(sqlPass string) {
	db := startSQL(sqlPass)
	defer db.Close()

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("failed to read cwd: %v", err)
	}

	kf := filepath.Join(workingDir, "keys")
	err = server.ReadKeys(kf)
	if err != nil {
		fmt.Printf("failed to read keyfile(%v) from disk: %v", "keys", err)
	}

	content, err := ioutil.ReadFile(kf)
	if err != nil {
		log.Errorln(err)
		return
	}

	// Fill the map with keys seen.
	for _, key := range strings.Split(string(content), ",") {
		keys[key] = true
	}

	n := 0
	for i := range keys {
		saltByte, _ := GenerateRandomBytes(40)
		salt := base64.URLEncoding.EncodeToString(saltByte)[:40]

		hash, err := bcrypt.GenerateFromPassword([]byte(string(string(alphabet[randInt()])+i+salt)), cost)
		if err != nil {
			panic(err)
		}

		fmt.Printf("user %v, key %v, salt: %v\n", n, string(hash), salt)
		n++
	}
}
