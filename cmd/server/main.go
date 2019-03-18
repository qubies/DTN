package main

import (
	"bytes"
	"encoding/gob"
	"errors"
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
	"text/tabwriter"
	"time"
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
	// if references[hash] <= 0 {
	//     os.Remove(filepath.Join(env.DATASTORE, hash))
	//     delete(references, hash)
	// }
}

func increaseCount(hash string) {
	refLock.Lock()
	defer refLock.Unlock()
	references[hash]++
}

func removeLinkCounts(name string, HOH string) error {
	f := filepath.Join(env.HASHLIST, name)
	if persist.FileExists(f) {
		fileRecord := persist.FileRecordFromFile(f)
		if _, ok := fileRecord.AllFiles[HOH]; ok {
			for _, hash := range fileRecord.AllFiles[HOH].Hashes {
				deleteCount(hash)
			}
			delete(fileRecord.AllFiles, HOH)
		}
		if len(fileRecord.AllFiles) < 1 {
			os.Remove(f)
		} else {
			if fileRecord.CurrentMainFile.HOH == HOH {
				var latest time.Time
				for _, record := range fileRecord.AllFiles {
					if record.ModifiedDate.Sub(latest) > 0 {
						fmt.Println("changed main")
						latest = record.ModifiedDate
						fileRecord.CurrentMainFile = record
					}

				}
			}
			fileRecord.Write()
		}

	} else {
		return errors.New("File not found")
	}
	return nil
}

func appendToFOB(name string, FileObject *persist.FileInfo) {
	f := filepath.Join(env.HASHLIST, name)
	var fileRecord *persist.FileRecord
	if persist.FileExists(f) {
		fileRecord = persist.FileRecordFromFile(f)
	} else {
		fileRecord = new(persist.FileRecord)
		fileRecord.AllFiles = make(map[string]*persist.FileInfo)
		fileRecord.FileName = name
	}
	fileRecord.CurrentMainFile = FileObject
	fileRecord.AllFiles[FileObject.HOH] = FileObject
	fileRecord.Write()

}

func uploadList(c *gin.Context) {
	fileName, _ := c.GetQuery("fileName")
	fileName = filepath.Base(fileName)
	hashData, _ := c.GetRawData()
	hashList := new(persist.FileInfo)
	dec := gob.NewDecoder(bytes.NewReader(hashData))
	err := dec.Decode(hashList)
	if err != nil {
		fmt.Println("Gob decode on uploaded list failed")
		c.String(http.StatusExpectationFailed, "HashList failed to decode")
		return
	}
	appendToFOB(fileName, hashList)
	c.String(200, "ok")
}

func getList(c *gin.Context) {
	//encode and send back the list!!
	fileName, _ := c.GetQuery("fileName")
	HOH, ok := c.GetQuery("HOH")
	target := filepath.Join(env.HASHLIST, fileName)
	fileExists := persist.FileExists(target)

	if !fileExists {
		logging.FileError("Filename Lookup", filepath.Join(env.HASHLIST, fileName), errors.New("uhoh"))
		c.String(404, "Not Found")
		return
	}

	FileRecord := persist.FileRecordFromFile(target)
	if ok {
		_, ok = FileRecord.AllFiles[HOH]
	}

	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	if ok {
		enc.Encode(FileRecord.AllFiles[HOH])
	} else {
		enc.Encode(FileRecord.CurrentMainFile)
	}
	c.Data(200, "binary/octet-stream", data.Bytes())
}

func loadRefs() {
	files, err := ioutil.ReadDir(env.HASHLIST)
	logging.PanicOnError("Unable to open hashlist directory", err)
	fmt.Println("Loading Hash List from store")
	for _, f := range files {
		thisMeta := persist.FileRecordFromFile(filepath.Join(env.HASHLIST, f.Name()))
		for _, hl := range thisMeta.AllFiles {
			for _, hash := range hl.Hashes {
				increaseCount(hash)
			}
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

func deleteFile(c *gin.Context) {
	fileName, _ := c.GetQuery("fileName")
	HOH, ok := c.GetQuery("HOH")
	if !ok || HOH == "" {
		c.String(200, "Remove Failed, you need to specify the HOH to remove")
	}
	err := removeLinkCounts(fileName, HOH)
	// err := os.Remove(filepath.Join(env.HASHLIST, fileName))
	if err == nil {
		c.String(200, "ok")
	} else {
		c.String(200, "Remove Failed")
	}
}

func fileList(c *gin.Context) {
	w := tabwriter.NewWriter(c.Writer, 0, 8, 0, ' ', tabwriter.Debug)
	files, err := ioutil.ReadDir(env.HASHLIST)
	logging.PanicOnError("Reading file list", err)
	fmt.Fprintln(w, " Name \t Size (bytes) \t Number of Blocks \t        Modified        \t   HOH    ")
	fmt.Fprintln(w, "------\t--------------\t------------------\t------------------------\t----------")
	for _, file := range files {
		fl := persist.FileRecordFromFile(filepath.Join(env.HASHLIST, file.Name()))
		for _, hl := range fl.AllFiles {
			fmt.Fprintf(w, " %v \t %v \t %v \t %v \t %v\n", file.Name(), hl.Size, len(hl.Hashes), hl.ModifiedDate.Format("Mon Jan _2 15:04:05 2006"), hl.HOH)
		}
	}
	w.Flush()
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
	router.GET("/fileList", fileList)
	router.GET("/DELETE", deleteFile)
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
