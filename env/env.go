package env

import (
	"fmt"
	e "github.com/joho/godotenv"
	hashing "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"os"
	"strconv"
)

//.env consts
var WD string
var DATASTORE string
var HASHLIST string
var BLOOMFILTER string
var LOGFILE string
var RESTPORT string
var SERVER_URL string
var BLOCK int

func BuildEnv() {
	e.Load(".env")
	WD = os.Getenv("WORKING_DIRECTORY")
	DATASTORE = os.Getenv("DATASTORE")
	HASHLIST = os.Getenv("HASH_LIST")
	BLOOMFILTER = os.Getenv("BLOOM_FILTER")
	LOGFILE = os.Getenv("LOGFILE")
	RESTPORT = os.Getenv("RESTPORT")
	SERVER_URL = os.Getenv("SERVER_URL")
	t, _ := strconv.Atoi(os.Getenv("BLOCK"))
	BLOCK = t
	persist.WD = WD
	logging.LOGFILE = LOGFILE
	hashing.BLOCK = BLOCK
	hashing.BLOCKSIZE = BLOCK * 1000

	//if the working directory does not exist, then create it.
	if _, err := os.Stat(WD); os.IsNotExist(err) {
		os.MkdirAll(WD, os.ModePerm)
		os.MkdirAll(DATASTORE, os.ModePerm)
		os.MkdirAll(HASHLIST, os.ModePerm)
	}
	fmt.Println("Working Directory:", WD)
}
