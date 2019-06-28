package main

import (
	"os"
	"path/filepath"

	"./server"

	"github.com/gidoBOSSftw5731/log"
)

func main() {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to read cwd: %v", err)
	}
	kf := filepath.Join(workingDir, keyFilename)
	err = server.ReadKeys(kf)
	if err != nil {
		log.Fatalf("failed to read keyfile(%v) from disk: %v", keyFilename, err)
	}
	
}
