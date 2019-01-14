package logging

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

const LOGLEVEL = log.InfoLevel

var LOGFILE string

func logStart() {
	log.WithFields(log.Fields{
		"task": "Logger initialized",
		"date": time.Now(),
	}).Info("Logger Initialized")
}

func Initialize() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(LOGLEVEL)
	logFile, err := os.OpenFile(LOGFILE, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic("Unable to open mailer logfile")
	}
	log.SetOutput(logFile)
	logStart()
}

func DuplicateFileWrite(fileName string) {
	task := "FileName written twice (already exists in DATA folder"
	log.WithFields(log.Fields{
		"task":     task,
		"fileName": fileName,
	}).Error(task)
}

func PanicObjectError(object interface{}, task string, err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"task":   task,
			"error":  err,
			"object": object,
		}).Error(task)
		panic(task)
	}
}

func main() {
	fmt.Println("vim-go")
}
