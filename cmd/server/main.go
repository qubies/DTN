package main

import (
	"fmt"
	env "github.com/joho/godotenv"
	hash "github.com/qubies/DTN/hashing"
	persist "github.com/qubies/DTN/persistentStore"
	log "github.com/sirupsen/logrus"
	"os"
)

var WD string
var HASHLIST string
var BLOOMFILTER string
var LOGFILE string

func buildEnv() {
	env.Load(".env")
	WD = os.Getenv("WORKING_DIRECTORY")
	HASHLIST = os.Getenv("HASH_LIST")
	BLOOMFILTER = os.Getenv("BLOOM_FILTER")
	LOGFILE = os.Getenv("LOGFILE")
	persist.WD = WD
}

func setupLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.WarnLevel)
	logFile, err := os.OpenFile(LOGFILE, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic("Unable to open mailer logfile")
	}
	log.SetOutput(logFile)
}

func main() {
	buildEnv()
	fmt.Println("Working Directory:", WD)

	//if the working directory does not exist, then create it.
	if _, err := os.Stat(WD); os.IsNotExist(err) {
		os.Mkdir(WD, os.ModePerm)
	}

	// curently this just generates a hashlist for testing purposes.
	hashes := hash.GenerateHashList("testfile")

	//build the persistent read write channels.
	readChan, readResponseChan, writeChan, writeResponseChan := persist.PersistentChannels()
	hashStore := persist.NewFOB(HASHLIST)

	hashStore.Object = hashes

	// send in the persistent Store request
	writeChan <- hashStore
	// wait for and remove the response
	<-writeResponseChan
	test := persist.NewFOB(HASHLIST)

	// send in the read request
	readChan <- test
	//block for response
	<-readResponseChan
}
