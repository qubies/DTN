package Hashing

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
)

const BLOCKSIZE int = 100 * 1000000
const NUM_WORKERS int = 8

func sha2Block(info []byte) [32]byte {
	return sha256.Sum256(info)
}

type OrderedReturn struct {
	index int
	bytes []byte
	hash  [32]byte
}

func workChan(input <-chan (*OrderedReturn), output chan<- (*OrderedReturn), wg *sync.WaitGroup) {
	for local := range input {
		local.hash = sha2Block(local.bytes)
		output <- local
	}
	wg.Done()
}

func consumer(input chan (*OrderedReturn), wg *sync.WaitGroup) [][32]byte {
	results := make(map[int][32]byte)
	for OR := range input {
		results[OR.index] = OR.hash
	}

	rVal := make([][32]byte, len(results))
	for k, v := range results {
		rVal[k] = v
	}
	defer wg.Done()
	return rVal
}

//this is the only exported function, it should generate a lsit of hashes.
func GenerateHashList(fileName string) [][32]byte {
	var results [][32]byte
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	dataChannel := make(chan *OrderedReturn, NUM_WORKERS)
	hashChannel := make(chan *OrderedReturn, NUM_WORKERS)

	wg := new(sync.WaitGroup)
	wg2 := new(sync.WaitGroup)
	wg.Add(NUM_WORKERS)
	wg2.Add(1)

	for i := 0; i < NUM_WORKERS; i++ {
		go workChan(dataChannel, hashChannel, wg)
	}

	index := 0

	go func() { results = consumer(hashChannel, wg2) }()

	for {
		OR := new(OrderedReturn)
		OR.index = index
		OR.bytes = make([]byte, BLOCKSIZE)
		if _, err := file.Read(OR.bytes); err == io.EOF {
			break
		}
		dataChannel <- OR
		index++
	}

	close(dataChannel)
	wg.Wait()
	close(hashChannel)
	wg2.Wait()

	// fmt.Println("Results: ", results)
	return results
}
