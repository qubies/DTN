package main

import (
	"fmt"
	env "github.com/joho/godotenv"
	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"os"

	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

//.env consts
var WD string
var DATASTORE string
var HASHLIST string
var BLOOMFILTER string
var LOGFILE string
var RESTPORT string

const MEMORYLIMIT = 200

func buildEnv() {
	env.Load(".env")
	WD = os.Getenv("WORKING_DIRECTORY")
	DATASTORE = os.Getenv("DATASTORE")
	HASHLIST = os.Getenv("HASH_LIST")
	BLOOMFILTER = os.Getenv("BLOOM_FILTER")
	LOGFILE = os.Getenv("LOGFILE")
	RESTPORT = os.Getenv("RESTPORT")
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

func runServer() {
	// adapted from the gin docs example

	//initialize the api
	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = MEMORYLIMIT << 20 // MEMORYLIMIT MiB
	router.Static("/", "./public")
	router.POST("/deposit", func(c *gin.Context) {
		fileName := c.PostForm("name")
		file, err := c.FormFile("file")

		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
		filename := filepath.Join(DATASTORE, file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully with fields name=%s ", file.Filename, fileName))
	})
	router.Run(":" + RESTPORT)
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
	runServer()
}
