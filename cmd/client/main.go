package main

import (
	"bytes"
	"fmt"
	// "fmt"
	"encoding/gob"
	env "github.com/qubies/DTN/env"
	hash "github.com/qubies/DTN/hashing"
	input "github.com/qubies/DTN/input"
	logging "github.com/qubies/DTN/logging"
	// persist "github.com/qubies/DTN/persistentStore"
	"io/ioutil"
	"net/http"
	"sync"
)

const num_senders = 100

func readResponse(response *http.Response) string {
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	logging.PanicOnError("reading get request body from checker", err)
	return string(contents)
}

func send(hash string, data []byte) bool {
	resp, err := http.Post("http://"+env.SERVER_URL+":"+env.RESTPORT+"/deposit?hash="+hash, "binary/octet-stream", bytes.NewReader(data))
	logging.PanicOnError("Error creating HTTP request", err)
	return readResponse(resp) == "ok"
}

func check(hash string) bool {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/check?hash=" + hash)
	logging.PanicOnError("Get Request to checker", err)
	return readResponse(response) == "SEND"
}
func sendHashList(fileName string, data *bytes.Buffer) bool {
	resp, err := http.Post("http://"+env.SERVER_URL+":"+env.RESTPORT+"/hashlist?fileName="+fileName, "binary/octet-stream", bytes.NewReader(data.Bytes()))
	logging.PanicOnError("Get Request to checker", err)
	return readResponse(resp) == "ok"
}

func main() {
	fileName := input.GetFile()
	env.BuildEnv()
	logging.Initialize()

	// curently this just generates a hashlist for testing purposes.
	fmt.Println("Workers On Sending Pipeline:", num_senders)
	partChan := hash.GenerateHashList(fileName)
	var wg sync.WaitGroup

	wg.Add(num_senders) //number of senders (warning that this will unplug the pipeline to a degree and use more memory)
	maxIndex := 0

	var hashList sync.Map
	var uniqueHash sync.Map
	for x := 0; x < num_senders; x++ {
		go func() {
			for x := range partChan {
				if x.Index > maxIndex {
					maxIndex = x.Index
				}
				hashList.Store(x.Index, x.Hash)
				_, ok := uniqueHash.LoadOrStore(x.Hash, true)
				if !ok {

					if check(x.Hash) {
						send(x.Hash, x.Bytes)
						// fmt.Println("Sent Hash:", x.Hash)
					}
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	finalList := make([]string, maxIndex+1)
	hashList.Range(func(key, value interface{}) bool {
		finalList[key.(int)] = value.(string)
		return true
	})
	var listStore bytes.Buffer
	enc := gob.NewEncoder(&listStore)
	enc.Encode(finalList)
	if sendHashList(fileName, &listStore) {
		fmt.Println("File Stored")
	}
	//build the persistent read write channels.
	// hashStore := persist.NewFOB(env.HASHLIST, finalList)
	// // hashStore.Object = finalList

	// // persistently write and ensure file is on drive
	// hashStore.WriteBlocking()

	// test := persist.NewFOB(env.HASHLIST, finalL)
	// test.ReadBlocking()
}
