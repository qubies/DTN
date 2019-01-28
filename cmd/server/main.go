package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/gin-gonic/gin"
	env "github.com/qubies/DTN/env"
	hashing "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var references map[string]int
var refLock sync.Mutex

func uploadPost(c *gin.Context) {
	fileName, _ := c.GetQuery("hash")
	data, _ := c.GetRawData()

	hashForThisBlock := hashing.HashBlock(data)
	if hashForThisBlock != fileName {
		fmt.Println("Hash did not match")
		c.String(http.StatusResetContent, "File Upload Incomplete")
	} else {
		persist.WriteBytes(filepath.Join(env.DATASTORE, fileName), data)
		c.String(200, "ok")
	}
}

func deleteCount(hash string) {
	refLock.Lock()
	defer refLock.Unlock()
	references[hash]--
	if references[hash] <= 0 {
		os.Remove(filepath.Join(env.DATASTORE, hash))
		delete(references, hash)
	}
}

func increaseCount(hash string) {
	refLock.Lock()
	defer refLock.Unlock()
	references[hash]++
}

func checkMeta(name string, newMeta *[]string) {
	f := filepath.Join(env.HASHLIST, name)
	if persist.FileExists(f) {
		oldHashList := persist.HashListFromFile(f)
		for _, hash := range *oldHashList {
			deleteCount(hash)
		}
	}
}

func uploadList(c *gin.Context) {
	fileName, _ := c.GetQuery("fileName")
	hashData, _ := c.GetRawData()
	hashList := new([]string)
	dec := gob.NewDecoder(bytes.NewReader(hashData))
	err := dec.Decode(hashList)
	if err != nil {
		fmt.Println("Gob decode on uploaded list failed")
		c.String(http.StatusExpectationFailed, "HashList failed to decode")
		return
	}
	checkMeta(fileName, hashList)
	persist.WriteBytes(filepath.Join(env.HASHLIST, fileName), hashData)
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

func loadRefs() {
	files, err := ioutil.ReadDir(env.HASHLIST)
	logging.PanicOnError("Unable to open hashlist directory", err)
	fmt.Println("Loading Hash List from store")
	for _, f := range files {
		thisMeta := persist.HashListFromFile(filepath.Join(env.HASHLIST, f.Name()))
		for _, hash := range *thisMeta {
			increaseCount(hash)
		}
	}
	fmt.Println("Cleaning up any unneeded files...")
	hashes, err := ioutil.ReadDir(env.DATASTORE)
	logging.PanicOnError("Unable to open hash directory", err)
	for _, hash := range hashes {
		if _, ok := references[hash.Name()]; !ok {
			os.Remove(filepath.Join(env.DATASTORE, hash.Name()))
		}
	}
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
	increaseCount(hash)
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
	references = make(map[string]int)
	env.BuildEnv()
	logging.Initialize()
	loadRefs()
}

func main() {
	startup()
	runServer()
}
