package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http/fcgi"
	"os"
	"path"
	"path/filepath"
	"time"

	hasholdkeys "./scripts/HashOldKeys"
	"./server"

	"github.com/gidoBOSSftw5731/log"
	_ "github.com/go-sql-driver/mysql"
)

// createImgDir creates all image storage directories
func createImgDir(imgStore string) {
	for f := 0; f < 16; f++ {
		for s := 0; s < 16; s++ {
			os.MkdirAll(filepath.Join(imgStore, fmt.Sprintf("%x/%x", f, s)), 0755)
		}
	}
	fmt.Println("Finished making/Verifying the directories!")
}

func logger() error {
	log.SetCallDepth(loggingLevel)
	switch loggingLevel {
	case 0:
		log.EnableLevel("fatal")
	case 1:
		log.EnableLevel("fatal")
		log.EnableLevel("error")
	case 2:
		log.EnableLevel("fatal")
		log.EnableLevel("error")
		log.EnableLevel("info")
	case 3:
		log.EnableLevel("fatal")
		log.EnableLevel("error")
		log.EnableLevel("info")
		log.EnableLevel("debug")
	case 4:
		log.EnableLevel("fatal")
		log.EnableLevel("error")
		log.EnableLevel("info")
		log.EnableLevel("debug")
		log.EnableLevel("trace")
	}
	//Set logging path
	logPath := path.Join("log/" + time.Now().Format("20060102"))
	logLatestPath := path.Join("log/" + "latest")
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("couldnt open Time dependent logfile(%v): %v", logPath, err)
	}
	defer logFile.Close()

	if _, err := os.Stat(logLatestPath); err == nil {
		err = os.Remove(logLatestPath)
	}

	if err != nil {
		return fmt.Errorf("Couldnt remove latest log file(%v) even though we didnt see it: %v", logLatestPath, err)
	}
	logLatest, err := os.OpenFile(logLatestPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("couldnt open non-Time dependent logfile(%v): %v", logLatestPath, err)
	}
	defer logLatest.Close()

	mw := io.MultiWriter(os.Stdout, logFile, logLatest)
	log.SetOutput(mw)
	return nil
}

func isFlagPassed() {
	legacykeys := flag.String("fixkeys", "", "correct legacy key system")
	fmt.Println(*legacykeys)
	flag.Parse()

	found := false
	flag.Visit(func(f *flag.Flag) {
		if *legacykeys != "" {
			found = true
			fmt.Println(found)
			hasholdkeys.Run(sqlPasswd)
			os.Exit(0)
		}
	})

}

//When everything gets set up, all page setup above this
func main() {
	isFlagPassed()
	go createImgDir(imgStore)

	fmt.Println("Starting the program.")
	listener, _ := net.Listen("tcp", "127.0.0.1:9001")
	fmt.Println("Started the listener.")
	srv := server.NewFastCGIServer(urlPrefix, imgStore, baseURL, sqlPasswd, recaptchaPrivKey, recaptchaPubKey, imgHash)
	fmt.Println("Starting the fcgi.")

	// I reccomend blocking 3306 in your firewall unless you use the port elsewhere
	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ImgSrvr", sqlPasswd))
	if err != nil {
		fmt.Println("Oh noez, could not connect to database")
		return
	}
	defer db.Close()
	fmt.Println("Oi, mysql did thing")

	//Enable logging
	err = logger()
	if err != nil {
		log.Fatalf("logging setup failed: %v", err)
	}

	//Debug:
	//This prints stuff in the console so i get info, just for me
	dir, err := os.Getwd()
	if err != nil {
		log.Error("Error happened!!! Here, take it: %v", err)
	}
	log.Debugf("DIR: %v\n", dir)
	log.Debugf("Heres the prefix for the url, dummy: %s \n", urlPrefix)
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
