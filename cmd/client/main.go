package main

import (
	"bytes"
	"fmt"
	// "fmt"
	"encoding/gob"
	env "github.com/qubies/DTN/env"
	hash "github.com/qubies/DTN/hashing"
	// "os"
	// hashing "github.com/qubies/DTN/hashing"
	input "github.com/qubies/DTN/input"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"io/ioutil"
	"net/http"
	"path/filepath"
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
func getHashList(fileName string) *[]string {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/getList?fileName=" + fileName)
	logging.PanicOnError("Get Request Hash List", err)
	hashList := new([]string)
	dec := gob.NewDecoder(response.Body)
	dec.Decode(hashList)
	response.Body.Close()
	return hashList
}

func getHash(hash string) *[]byte {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/getData?hash=" + hash)
	logging.PanicOnError("Get Request Hash", err)
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	return &b
}

func main() {
	fileName, op := input.CollectOptions()
	env.BuildEnv()
	logging.Initialize()

	if op == 'u' {

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
	} else if op == 'd' {
		//recreate the file for a test to ./rebuilt.
		hashList := getHashList(fileName)
		for _, x := range *hashList {
			d := getHash(x)
			persist.WriteBytes(filepath.Join(env.DATASTORE, x), *d)
		}
		fmt.Println(hashList)
		hash.Rebuild(hashList, env.DATASTORE, fileName+".rebuilt")
	}
}
