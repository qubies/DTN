package Hashing

import (
	hashFunc "crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"time"
)

const BLOCKSIZE int = 1 * 1000000

var NUM_WORKERS int = runtime.GOMAXPROCS(0)

var BUFFERSIZE int = NUM_WORKERS + 1 // add some space for the gatherer

func hashBlock(info []byte) string {
	h := hashFunc.New()
	h.Write(info)
	return hex.EncodeToString(h.Sum(nil))
}

type FilePart struct {
	Index int
	Bytes []byte
	Hash  string
}

type indexCount struct {
	sync.RWMutex
	val int
}

// The index count prevents the workers from filling the buffer poorly and also from overstuffing it.
func (I *indexCount) incr() {
	I.Lock()
	defer I.Unlock()
	I.val++
}

func (I *indexCount) get() int {
	I.RLock()
	defer I.RUnlock()
	return I.val
}

func workChan(input <-chan (*FilePart), output chan<- (*FilePart), wg *sync.WaitGroup, ic *indexCount) {
	for local := range input {
		local.Hash = hashBlock(local.Bytes)
		for local.Index >= ic.get()+BUFFERSIZE-1 {
			// fmt.Println("blocked")
			time.Sleep(time.Millisecond * 1)
		}
		output <- local
	}
	wg.Done()
}

func gatherer(input chan (*FilePart), indexChan chan int, ic *indexCount) chan *FilePart {
	output := make(chan *FilePart, BUFFERSIZE)
	go func() {

		cVal := 0
		index := 999999999999
		for cVal < index {
			select {
			case OR := <-input:
				if OR.Index != cVal {
					input <- OR
				} else {
					cVal++
					output <- OR
					ic.incr()
				}
			case index = <-indexChan:
				{
				}
			}
		}

		defer close(output)
	}()
	return output
}

//this is the only exported function, it should generate a lsit of Hashes.
func GenerateHashList(fileName string) chan *FilePart {
	fmt.Println("Workers:", NUM_WORKERS)
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	dataChannel := make(chan *FilePart, NUM_WORKERS)
	HashChannel := make(chan *FilePart, NUM_WORKERS)
	indexChannel := make(chan int)

	wg := new(sync.WaitGroup)
	wg.Add(NUM_WORKERS)
	ic := new(indexCount)
	for i := 0; i < NUM_WORKERS; i++ {
		go workChan(dataChannel, HashChannel, wg, ic)
	}

	Index := 0
	output := gatherer(HashChannel, indexChannel, ic)

	go func() {
		for {
			OR := new(FilePart)
			OR.Index = Index
			OR.Bytes = make([]byte, BLOCKSIZE)
			if _, err := file.Read(OR.Bytes); err == io.EOF {
				break
			}
			dataChannel <- OR
			Index++
		}
		indexChannel <- Index

		close(dataChannel)
		wg.Wait()
	}()
	return output
}

func HashFile(fileName string) string {
	f, _ := os.Open(fileName)
	d, _ := ioutil.ReadAll(f)
	fmt.Println("FileHash:", hashBlock(d))
	return hashBlock(d)
}
