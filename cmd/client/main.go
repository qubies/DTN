package main

import (
	"bytes"
	"fmt"
	env "github.com/qubies/DTN/env"
	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"io/ioutil"
	"net/http"
	"sync"
)

const num_senders = 10

func send(hash string, data []byte) {
	resp, err := http.Post("http://"+env.SERVER_URL+":"+env.RESTPORT+"/deposit?hash="+hash, "binary/octet-stream", bytes.NewReader(data))
	logging.PanicOnError("Error creating HTTP request", err)
	resp.Body.Close()
}

func check(hash string) bool {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/check?hash=" + hash)
	logging.PanicOnError("Get Request to checker", err)
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	logging.PanicOnError("reading get request body from checker", err)
	return string(contents) == "SEND"
}

func main() {
	env.BuildEnv()
	logging.Initialize()

	// curently this just generates a hashlist for testing purposes.
	hl := new([]string)
	partChan := hash.GenerateHashList("testfile")
	var wg sync.WaitGroup
	wg.Add(num_senders) //number of senders (warning that this will unplug the pipeline to a degree and use more memory)
	var hashList []string
	for x := 0; x < num_senders; x++ {
		go func() {
			for x := range partChan {
				hashList = append(hashList, x.Hash)
				if check(x.Hash) {
					send(x.Hash, x.Bytes)
					fmt.Println("Sent Hash:", x.Hash)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	//build the persistent read write channels.
	hashStore := persist.NewFOB(env.HASHLIST, hl)
	hashStore.Object = hashList

	// persistently write and ensure file is on drive
	hashStore.WriteBlocking()

	test := persist.NewFOB(env.HASHLIST, hl)
	test.ReadBlocking()
}
