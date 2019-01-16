package env

import (
	"fmt"
	e "github.com/joho/godotenv"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"os"
)

//.env consts
var WD string
var DATASTORE string
var HASHLIST string
var BLOOMFILTER string
var LOGFILE string
var RESTPORT string

func BuildEnv() {
	e.Load(".env")
	WD = os.Getenv("WORKING_DIRECTORY")
	DATASTORE = os.Getenv("DATASTORE")
	HASHLIST = os.Getenv("HASH_LIST")
	BLOOMFILTER = os.Getenv("BLOOM_FILTER")
	LOGFILE = os.Getenv("LOGFILE")
	RESTPORT = os.Getenv("RESTPORT")
	persist.WD = WD
	logging.LOGFILE = LOGFILE

	//if the working directory does not exist, then create it.
	if _, err := os.Stat(WD); os.IsNotExist(err) {
		os.MkdirAll(WD, os.ModePerm)
		os.MkdirAll(DATASTORE, os.ModePerm)
	}

	fmt.Println("Working Directory:", WD)
}
