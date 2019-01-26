package Hashing

import (
	hashFunc "crypto/sha256"
	"encoding/hex"
	"fmt"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var BLOCK int
var BLOCKSIZE int

var NUM_WORKERS int = runtime.GOMAXPROCS(0)

// var NUM_WORKERS int = 1

var BUFFERSIZE int = NUM_WORKERS + 1 // add some space for the gatherer

func HashBlock(info []byte) string {
	h := hashFunc.New()
	h.Write(info)
	return hex.EncodeToString(h.Sum(nil))
}

type FilePart struct {
	Index int
	Bytes []byte
	Hash  string
}

func workChan(input <-chan (*FilePart), output chan<- (*FilePart), wg *sync.WaitGroup) {
	for local := range input {
		local.Hash = HashBlock(local.Bytes)
		output <- local
	}
	wg.Done()
}

//this is the only exported function, it should generate a lsit of Hashes.
func GenerateHashList(fileName string) (chan *FilePart, *pb.ProgressBar) {
	fmt.Println("Workers On Hash Pipeline:", NUM_WORKERS)
	fmt.Println("Blocksize:", BLOCKSIZE/1000, "kB")
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	dataChannel := make(chan *FilePart, NUM_WORKERS)
	hashChannel := make(chan *FilePart, NUM_WORKERS)

	wg := new(sync.WaitGroup)
	wg.Add(NUM_WORKERS)
	for i := 0; i < NUM_WORKERS; i++ {
		go workChan(dataChannel, hashChannel, wg)
	}
	Index := 0
	fileInfo, err := file.Stat()
	if err != nil {
		// Could not obtain stat, handle error
	}

	bar := pb.StartNew(int(fileInfo.Size()/int64(BLOCKSIZE) + 1))
	go func() {
		for {
			OR := new(FilePart)
			OR.Index = Index
			OR.Bytes = make([]byte, BLOCKSIZE)
			bytesRead, err := file.Read(OR.Bytes)
			if err == io.EOF {
				break
			}
			if bytesRead != BLOCKSIZE {
				OR.Bytes = append([]byte(nil), OR.Bytes[:bytesRead]...)
			}
			dataChannel <- OR
			// bar.Increment()
			Index++
		}

		close(dataChannel)
		wg.Wait()
		close(hashChannel)
		// bar.FinishPrint("The End!")
	}()
	return hashChannel, bar
}

func Rebuild(hashList *[]string, directory string, finalPath string) {
	// hashList := persist.HashListFromFile(filePath)

	output, err := os.Create(finalPath)
	if err != nil {
		panic("Error creating file:" + err.Error())
	}
	defer output.Close()
	bar := pb.StartNew(len(*hashList))
	for _, x := range *hashList {
		p, err := os.Open(filepath.Join(directory, x))
		if err != nil {
			panic("Error opening file:" + err.Error())
		}
		io.Copy(output, p)
		bar.Add(1)
	}
	bar.FinishPrint("Rebuild Complete!")
}

func HashFile(fileName string) string {
	f, _ := os.Open(fileName)
	d, _ := ioutil.ReadAll(f)
	fmt.Println("FileHash:", HashBlock(d))
	return HashBlock(d)
}
