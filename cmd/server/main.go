package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	env "github.com/qubies/DTN/env"
	//	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	//	persist "github.com/qubies/DTN/persistentStore"
	"net/http"
	"path/filepath"
)

const MEMORYLIMIT = 200

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
		filename := filepath.Join(env.DATASTORE, file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully with fields name=%s ", file.Filename, fileName))
	})
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
