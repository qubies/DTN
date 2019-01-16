package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	env "github.com/qubies/DTN/env"
	//	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	//	persist "github.com/qubies/DTN/persistentStore"
	"net/http"
	"os"
	"path/filepath"
)

func uploadPost(c *gin.Context) {
	fileName, _ := c.GetQuery("hash")
	data, _ := c.GetRawData()
	persist.WriteBytes(filepath.Join(env.DATASTORE, fileName), data)
	c.String(200, "ok")
}

func checkHash(c *gin.Context) {
	hash, _ := c.GetQuery("hash")
	fmt.Println("Checking for Hash:", hash)
	if _, err := os.Stat(filepath.Join(env.DATASTORE, hash)); os.IsNotExist(err) {
		c.String(http.StatusOK, "SEND")
	} else {
		c.String(http.StatusOK, "NO-SEND")
	}
}

func runServer() {
	// adapted from the gin docs example
	//initialize the api
	router := gin.Default()

	router.POST("/deposit", uploadPost)
	router.GET("/check", checkHash)
	router.Run(":" + env.RESTPORT)
}
func startup() {
	env.BuildEnv()
	logging.Initialize()
}

func main() {
	startup()
	runServer()
}
