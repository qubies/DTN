package env

import (
	"fmt"
	e "github.com/joho/godotenv"
	hashing "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"os"
	"path/filepath"
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

var NUM_DOWNLOAD_WORKERS int
var NUM_UPLOAD_WORKERS int
var NUM_HASH_WORKERS int

func BuildEnv() {
	e.Load(".env")
	WD = os.Getenv("WORKING_DIRECTORY")
	DATASTORE = os.Getenv("DATASTORE")
	HASHLIST = os.Getenv("HASH_LIST")
	BLOOMFILTER = os.Getenv("BLOOM_FILTER")
	LOGFILE = os.Getenv("LOGFILE")
	RESTPORT = os.Getenv("RESTPORT")
	SERVER_URL = os.Getenv("SERVER_URL")
	NUM_DOWNLOAD_WORKERS, _ = strconv.Atoi(os.Getenv("NUM_DOWNLOAD_WORKERS"))
	NUM_UPLOAD_WORKERS, _ = strconv.Atoi(os.Getenv("NUM_UPLOAD_WORKERS"))
	NUM_HASH_WORKERS, _ = strconv.Atoi(os.Getenv("NUM_HASH_WORKERS"))
	BLOCK, _ = strconv.Atoi(os.Getenv("BLOCK"))
	persist.WD = WD
	logging.LOGFILE = LOGFILE
	hashing.BLOCK = BLOCK
	hashing.BLOCKSIZE = BLOCK * 1000
	hashing.NUM_WORKERS = NUM_HASH_WORKERS

	//if the working directory does not exist, then create it.
	if _, err := os.Stat(WD); os.IsNotExist(err) {
		os.MkdirAll(WD, os.ModePerm)
		os.MkdirAll(DATASTORE, os.ModePerm)
		os.MkdirAll(HASHLIST, os.ModePerm)
		os.MkdirAll(filepath.Join(WD, "tmp"), os.ModePerm)
	}
	fmt.Println("Working Directory:", WD)
}
