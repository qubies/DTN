package persistentStore

import (
	"encoding/gob"
	log "github.com/sirupsen/logrus"
	"os"
)

type FileObject struct {
	filePath string
	object   interface{}
	confirm  bool
}

var wd string

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
	tmp_name := wd + "/tmp"
	tmp, err := os.Create(tmp_name)
	fobErr(fob, "Error creating temp file for object in writeFob", err)

	encoder := gob.NewEncoder(tmp)
	err = encoder.Encode(fob.object)
	fobErr(fob, "Error encoding temp file for object in writeFob", err)

	tmp.Close()
	err = os.Rename(tmp_name, fob.filePath)
	fobErr(fob, "Error renaming temp file for object in writeFob", err)
}

//https://medium.com/@kpbird/golang-serialize-struct-using-gob-part-1-e927a6547c00
func readGob(fob *FileObject) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}

func Create_Persistent_Objects() {
	//write_channel
	//read_request_channel
	//write_confirmation_channel

}
