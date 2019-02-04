package Hashing

import (
	"bufio"
	hashFunc "crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cespare/xxhash"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var DYNAMIC bool
var HASH_WINDOW_SIZE int
var HASH_MATCHING_STRING string
var MINIMUM_BLOCK_SIZE int
var MAXIMUM_BLOCK_SIZE int

var NUM_WORKERS int

var BUFFERSIZE int = NUM_WORKERS + 1 // add some space for the gatherer

func HashBlock(info []byte) string {
	h := hashFunc.New()
	h.Write(info)
	return hex.EncodeToString(h.Sum(nil))
}

func xx(info []byte) string {
	// fmt.Println("xxhash:", xxhash.Sum64(info))
	return fmt.Sprintf("%20d", xxhash.Sum64(info))
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

	bar := pb.StartNew(int(fileInfo.Size())).SetUnits(pb.U_BYTES)
	bar.Start()

	go func() {
		rd := bufio.NewReader(file)
		for {
			bytesRead := 0
			OR := new(FilePart)
			OR.Index = Index
			OR.Bytes = make([]byte, MAXIMUM_BLOCK_SIZE)
			var err error
			var c byte
			for {
				if bytesRead == MAXIMUM_BLOCK_SIZE {
					// fmt.Println("The block was filled")
					break
				}
				c, err = rd.ReadByte()
				if err != nil {
					break
				}
				OR.Bytes[bytesRead] = byte(c)
				if DYNAMIC {

					hv := xx(OR.Bytes[max(bytesRead-HASH_WINDOW_SIZE-1, 0):bytesRead])
					bytesRead++
					if hv[len(hv)-len(HASH_MATCHING_STRING):] == HASH_MATCHING_STRING {
						if bytesRead < MINIMUM_BLOCK_SIZE {
							continue
						}
						// fmt.Println("Match")
						break
					}
				} else {
					bytesRead++
				}
			}
			if bytesRead != MAXIMUM_BLOCK_SIZE {
				OR.Bytes = append([]byte(nil), OR.Bytes[:bytesRead]...)
			}
			dataChannel <- OR
			//this is a little tenuous.
			if err != nil {
				break
			}
			Index++
		}

		close(dataChannel)
		wg.Wait()
		close(hashChannel)
		// bar.FinishPrint("The End!")
	}()
	return hashChannel, bar
}

func max(a, b int) int {
	if a <= b {
		return b
	}
	return a
}

func Rebuild(hashList *[]string, directory string, finalPath string) {
	// hashList := persist.HashListFromFile(filePath)

	output, err := os.Create(finalPath)
	if err != nil {
		panic("Error creating file:" + err.Error())
	}
	defer output.Close()
	bar := pb.StartNew(len(*hashList))
	bar.Start()
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
