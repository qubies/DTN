package env

import (
	e "github.com/joho/godotenv"
	hashing "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//.env consts
var WD string
var DATASTORE string
var HASHLIST string
var BLOOMFILTER string
var LOGFILE string
var RESTPORT string
var SERVER_URL string
var DYNAMIC bool

var NUM_DOWNLOAD_WORKERS int
var NUM_UPLOAD_WORKERS int
var NUM_HASH_WORKERS int

var MINIMUM_BLOCK_SIZE int
var MAXIMUM_BLOCK_SIZE int
var HASH_WINDOW_SIZE int
var HASH_MATCHING_STRING string

func BuildEnv() {
	e.Load(".env")

	WD = os.Getenv("WORKING_DIRECTORY")
	DATASTORE = os.Getenv("DATASTORE")
	HASHLIST = os.Getenv("HASH_LIST")
	BLOOMFILTER = os.Getenv("BLOOM_FILTER")
	LOGFILE = os.Getenv("LOGFILE")
	RESTPORT = os.Getenv("RESTPORT")
	SERVER_URL = os.Getenv("SERVER_URL")
	HASH_MATCHING_STRING = os.Getenv("HASH_MATCHING_STRING")
	DYNAMIC = strings.ToLower(os.Getenv("DYNAMIC")) == "true"

	NUM_DOWNLOAD_WORKERS, _ = strconv.Atoi(os.Getenv("NUM_DOWNLOAD_WORKERS"))
	NUM_UPLOAD_WORKERS, _ = strconv.Atoi(os.Getenv("NUM_UPLOAD_WORKERS"))
	NUM_HASH_WORKERS, _ = strconv.Atoi(os.Getenv("NUM_HASH_WORKERS"))
	MINIMUM_BLOCK_SIZE, _ = strconv.Atoi(os.Getenv("MINIMUM_BLOCK_SIZE"))
	MAXIMUM_BLOCK_SIZE, _ = strconv.Atoi(os.Getenv("MAXIMUM_BLOCK_SIZE"))
	HASH_WINDOW_SIZE, _ = strconv.Atoi(os.Getenv("HASH_WINDOW_SIZE"))

	persist.WD = WD
	persist.HASH_STORAGE = HASHLIST
	logging.LOGFILE = LOGFILE
	hashing.NUM_WORKERS = NUM_HASH_WORKERS
	hashing.HASH_WINDOW_SIZE = HASH_WINDOW_SIZE
	hashing.HASH_MATCHING_STRING = HASH_MATCHING_STRING
	hashing.MINIMUM_BLOCK_SIZE = MINIMUM_BLOCK_SIZE
	hashing.MAXIMUM_BLOCK_SIZE = MAXIMUM_BLOCK_SIZE
	hashing.DYNAMIC = DYNAMIC

	//if the working directory does not exist, then create it.
	if _, err := os.Stat(WD); os.IsNotExist(err) {
		os.MkdirAll(WD, os.ModePerm)
		os.MkdirAll(DATASTORE, os.ModePerm)
		os.MkdirAll(HASHLIST, os.ModePerm)
		os.MkdirAll(filepath.Join(WD, "tmp"), os.ModePerm)
	}
	// fmt.Println("Working Directory:", WD)
}
