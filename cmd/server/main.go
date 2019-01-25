package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	env "github.com/qubies/DTN/env"
	//	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"net/http"
	"os"
	"path/filepath"
)

func uploadPost(c *gin.Context) {
	fileName, _ := c.GetQuery("hash")
	data, _ := c.GetRawData()
	// fmt.Println(fileName, data)
	persist.WriteBytes(filepath.Join(env.DATASTORE, fileName), data)
	c.String(200, "ok")
}

func uploadList(c *gin.Context) {
	fileName, _ := c.GetQuery("fileName")
	hashList, _ := c.GetRawData()
	// fmt.Println(fileName, data)
	persist.WriteBytes(filepath.Join(env.HASHLIST, fileName), hashList)
	c.String(200, "ok")

}

func getList(c *gin.Context) {
	//encode and send back the list!!
	fileName, _ := c.GetQuery("fileName")
	data, err := persist.ReadBytes(filepath.Join(env.HASHLIST, fileName))
	if err != nil {
		logging.FileError("Filename Lookup", filepath.Join(env.HASHLIST, fileName), err)
		c.String(404, "Not Found")
		return
	}
	c.Data(200, "binary/octet-stream", data)
}

func getData(c *gin.Context) {
	hash, _ := c.GetQuery("hash")
	data, err := persist.ReadBytes(filepath.Join(env.DATASTORE, hash))
	if err != nil {
		logging.FileError("Hash Lookup", filepath.Join(env.HASHLIST, hash), err)
		c.String(404, "Not Found")
		return
	}
	c.Data(200, "binary/octet-stream", data)
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
	// router := gin.New()

	router.POST("/deposit", uploadPost)
	router.GET("/check", checkHash)
	router.GET("/getList", getList)
	router.GET("/getData", getData)
	router.POST("/hashlist", uploadList)
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
