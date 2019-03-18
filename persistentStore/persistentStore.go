package persistentStore

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	logging "github.com/qubies/DTN/logging"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

var WD string
var HASH_STORAGE string
var tmpFileNum uint32

type FileInfo struct {
	Hashes       []string
	Size         uint64
	ModifiedDate time.Time
	HOH          string
}

func (F *FileRecord) Write() {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(F)
	WriteBytes(filepath.Join(HASH_STORAGE, F.FileName), buf.Bytes())
}

type FileRecord struct {
	FileName        string
	CurrentMainFile *FileInfo
	AllFiles        map[string]*FileInfo
}

// FileObject is a convenience wrapper to a persistent store object.
type FileWriter struct {
	FileName string
}

func FileExists(fileName string) bool {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return false
	}
	return true
}

func FileSize(fileName string) int64 {
	fi, err := os.Stat(fileName)
	if err != nil {
		return 0
	}
	return fi.Size()
}

func (f *FileWriter) Write(p []byte) (n int, err error) {
	tmpName := WD + "/tmp/" + fmt.Sprint(atomic.AddUint32(&tmpFileNum, 1))
	tmp, err := os.Create(tmpName)
	if err != nil {
		logging.FileError("Writing File", f.FileName, err)
		return 0, err
	}
	defer os.Remove(tmpName)

	r, err := tmp.Write(p)
	if err != nil {
		logging.FileError("Writing File", f.FileName, err)
		return 0, err
	}
	tmp.Sync()
	tmp.Close()
	err = os.Rename(tmpName, f.FileName)
	if err != nil {
		logging.FileError("Writing File", f.FileName, err)
		return 0, err
	}
	return r, nil
}

func WriteBytes(fileName string, data []byte) {
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		logging.DuplicateFileWrite(fileName)
	}
	fn := new(FileWriter)
	fn.FileName = fileName
	fn.Write(data)
}

func ReadBytes(fileName string) ([]byte, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, errors.New("File Did Not Exist")
	}
	return ioutil.ReadFile(fileName)
}

func FileRecordFromFile(filePath string) *FileRecord {
	tmp := new(FileRecord)
	gFile, err := os.Open(filePath)
	defer gFile.Close()
	if err != nil {
		logging.PanicOnError("Opening Gobfile", err)
	}
	dec := gob.NewDecoder(gFile)
	dec.Decode(tmp)
	return tmp
}
