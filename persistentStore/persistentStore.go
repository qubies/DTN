package persistentStore

import (
	"encoding/gob"
	logging "github.com/qubies/DTN/logging"
	// log "github.com/sirupsen/logrus"
	"errors"
	"io/ioutil"
	"os"
)

// FileObject is a convenience wrapper to a persistent store object.
type FileObject struct {
	FileName      string
	Object        interface{}
	Complete_Chan chan bool
}

var WD string
var DATASTORE string
var writeChan chan *FileObject
var readChan chan *FileObject

func fobErr(fob *FileObject, task string, err error) {
	logging.PanicObjectError(fob, task, err)
}

func writeFob(fob *FileObject) {
	tmpName := WD + "/tmp"
	tmp, err := os.Create(tmpName)
	fobErr(fob, "Error creating temp file for object in writeFob", err)

	encoder := gob.NewEncoder(tmp)
	err = encoder.Encode(fob.Object)
	fobErr(fob, "Error encoding temp file for object in writeFob", err)
	// not sure if nexessary for this project, This will slow down IO as the File buffers actually get flushed to disk.
	// The advantage is that once sync returns, the file is on the disk.
	tmp.Sync()
	tmp.Close()
	err = os.Rename(tmpName, fob.FileName)
	fobErr(fob, "Error renaming temp file for object in writeFob", err)
}

func readFob(fob *FileObject) {
	file, err := os.Open(fob.FileName)
	fobErr(fob, "Error loading file for object in readFob", err)
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(fob.Object)
	fobErr(fob, "Error decoding file for object in readFob", err)
}

func readRequests(fChan <-chan *FileObject) {
	for x := range fChan {
		readFob(x)
		x.Complete_Chan <- true
	}
}

func writeRequests(fChan <-chan *FileObject) {
	for x := range fChan {
		writeFob(x)
		x.Complete_Chan <- true
	}
}

// PersistentChannels return the handlers for the persistent fucntions.
func initialize() {
	//write_channel
	writeChan = make(chan *FileObject)
	//read_request_channel
	readChan = make(chan *FileObject)

	go readRequests(readChan)
	go writeRequests(writeChan)
}

// NewFOB returns a new fileObject with the assigned filename for persistent storage.
func NewFOB(fileName string, object interface{}) *FileObject {
	if writeChan == nil {
		initialize()
	}
	return &FileObject{FileName: fileName, Object: object, Complete_Chan: make(chan bool)}
}

func (FOB *FileObject) Write() {
	writeChan <- FOB
	go func() {
		<-FOB.Complete_Chan
	}()
}

func (FOB *FileObject) Read() {
	readChan <- FOB
	go func() {
		<-FOB.Complete_Chan
	}()
}
func (FOB *FileObject) WriteBlocking() {
	writeChan <- FOB
	<-FOB.Complete_Chan
}

func (FOB *FileObject) ReadBlocking() {
	readChan <- FOB
	<-FOB.Complete_Chan
}

func WriteBytes(fileName string, data []byte) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		file, err := os.Create(fileName)
		logging.PanicObjectError(file, "Creating new data file", err)
		defer file.Close()
		_, err = file.Write(data)
		logging.PanicObjectError(file, "Writing to new data file", err)
		file.Sync()
	} else {
		logging.DuplicateFileWrite(fileName)
	}
}

func ReadBytes(fileName string) ([]byte, error) {
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		return nil, errors.New("File Did Not Exist")
	}
	return ioutil.ReadFile(fileName)
}
