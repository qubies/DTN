package persistentStore

import (
	"encoding/gob"
	log "github.com/sirupsen/logrus"
	"os"
)

// FileObject is a convenience wrapper to a persistent store object.
type FileObject struct {
	FileName string
	Object   [][32]byte
}

var WD string

func fobErr(fob *FileObject, task string, err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"task":  task,
			"error": err,
			"fob":   fob,
		}).Error(task)
		panic(task)
	}
}

func writeFob(fob *FileObject) {
	tmpName := WD + "/tmp"
	tmp, err := os.Create(tmpName)
	fobErr(fob, "Error creating temp file for object in writeFob", err)

	encoder := gob.NewEncoder(tmp)
	err = encoder.Encode(fob.Object)
	fobErr(fob, "Error encoding temp file for object in writeFob", err)
	tmp.Close()
	err = os.Rename(tmpName, fob.FileName)
	fobErr(fob, "Error renaming temp file for object in writeFob", err)
}

func readFob(fob *FileObject) {
	file, err := os.Open(fob.FileName)
	fobErr(fob, "Error loading file for object in readFob", err)
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&fob.Object)
	fobErr(fob, "Error decoding file for object in readFob", err)
}

func readRequests(fChan <-chan *FileObject, rChan chan<- *FileObject) {
	for x := range fChan {
		readFob(x)
		rChan <- x
	}
}

func writeRequests(fChan <-chan *FileObject, rChan chan<- *FileObject) {
	for x := range fChan {
		writeFob(x)
		rChan <- x
	}
}

// PersistentChannels return the handlers for the persistent fucntions.
func PersistentChannels() (chan<- *FileObject, <-chan *FileObject, chan<- *FileObject, <-chan *FileObject) {
	//write_channel
	writeChan := make(chan *FileObject)
	//write_confirmation_channel
	writeConfirmChan := make(chan *FileObject)
	//read_request_channel
	readChan := make(chan *FileObject)
	//read_return_channel
	readResponseChan := make(chan *FileObject)
	go readRequests(readChan, readResponseChan)
	go writeRequests(writeChan, writeConfirmChan)
	return readChan, readResponseChan, writeChan, writeConfirmChan
}

// NewFOB returns a new fileObject with the assigned filename for persistent storage.
func NewFOB(fileName string) *FileObject {
	return &FileObject{FileName: fileName}
}
