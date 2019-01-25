package logging

import (
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var LOGFILE string

const LOGLEVEL = log.InfoLevel

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

func PanicOnError(task string, err error) {
	Error(task, err)
	panic(err)
}

func Error(task string, err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"task":  task,
			"error": err,
		}).Error(task)
	}
}

func FileError(task string, filePath string, err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"task":  task,
			"error": err,
			"file":  filePath,
		}).Error(task)
	}
}
