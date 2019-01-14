package main

import (
	"fmt"
	env "github.com/joho/godotenv"
	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"os"
)

var WD string
var DATASTORE string
var HASHLIST string
var BLOOMFILTER string
var LOGFILE string

func buildEnv() {
	env.Load(".env")
	WD = os.Getenv("WORKING_DIRECTORY")
	DATASTORE = os.Getenv("DATASTORE")
	HASHLIST = os.Getenv("HASH_LIST")
	BLOOMFILTER = os.Getenv("BLOOM_FILTER")
	LOGFILE = os.Getenv("LOGFILE")
	persist.WD = WD
	persist.DATASTORE = DATASTORE
	logging.LOGFILE = LOGFILE

	//if the working directory does not exist, then create it.
	if _, err := os.Stat(WD); os.IsNotExist(err) {
		os.MkdirAll(WD, os.ModePerm)
		os.MkdirAll(DATASTORE, os.ModePerm)
	}

	fmt.Println("Working Directory:", WD)
}

func main() {
	buildEnv()
	logging.Initialize()

	// curently this just generates a hashlist for testing purposes.
	hl := new([][32]byte)
	hashes := hash.GenerateHashList("testfile")

	//build the persistent read write channels.
	hashStore := persist.NewFOB(HASHLIST, hl)
	hashStore.Object = hashes

	// persistently write and ensure file is on drive
	hashStore.WriteBlocking()

	test := persist.NewFOB(HASHLIST, hl)
	test.ReadBlocking()
	fmt.Println("FOB:", test.Object.(*[][32]byte))
}
