package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	env "github.com/qubies/DTN/env"
	hashing "github.com/qubies/DTN/hashing"
	input "github.com/qubies/DTN/input"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"gopkg.in/cheggaaa/pb.v1"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

//number of senders (warning that this will unplug the pipeline to a degree and use more memory)

func readResponse(response *http.Response) string {
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	logging.PanicOnError("reading get request body from checker", err)
	return string(contents)
}

func sendFileBlock(hash string, data []byte) bool {
	resp, err := http.Post("http://"+env.SERVER_URL+":"+env.RESTPORT+"/deposit?hash="+hash, "binary/octet-stream", bytes.NewReader(data))
	logging.PanicOnError("Error creating HTTP request", err)
	return readResponse(resp) == "ok"
}

func getFileBlock(hash string) *[]byte {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/getData?hash=" + hash)
	logging.PanicOnError("Get Request Hash", err)
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	hashOfB := hashing.HashBlock(b)
	if hashOfB != hash {
		panic("downloaded hash does not match")
	}

	return &b
}

func checkForHashOnServer(hash string) bool {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/check?hash=" + hash)
	logging.PanicOnError("Get Request to checker", err)
	return readResponse(response) == "SEND"
}

func deleteFileFromServer(fileName string) bool {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/DELETE?fileName=" + fileName)
	logging.PanicOnError("DELETE request error", err)
	return readResponse(response) == "ok"
}

func sendHashList(fileName string, data *bytes.Buffer) bool {
	resp, err := http.Post("http://"+env.SERVER_URL+":"+env.RESTPORT+"/hashlist?fileName="+fileName, "binary/octet-stream", bytes.NewReader(data.Bytes()))
	logging.PanicOnError("Get Request to checker", err)
	return readResponse(resp) == "ok"
}

func getHashList(fileName string) *[]string {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/getList?fileName=" + fileName)
	logging.PanicOnError("Get Request Hash List", err)
	if response.StatusCode == http.StatusNotFound {
		fmt.Println("File Not Found on Server")
		os.Exit(1)
	}
	hashList := new([]string)
	dec := gob.NewDecoder(response.Body)
	dec.Decode(hashList)
	response.Body.Close()
	return hashList
}

func workDownloads(input chan string, wg *sync.WaitGroup, bar *pb.ProgressBar) {
	for x := range input {
		d := getFileBlock(x)
		persist.WriteBytes(filepath.Join(env.DATASTORE, x), *d)
		bar.Add(env.BLOCK * 1000)
	}
	wg.Done()
}

func upload(fileName string) {
	fmt.Println("Workers On Sending Pipeline:", env.NUM_UPLOAD_WORKERS)

	fileBlockChannel, bar := hashing.GenerateHashList(fileName)

	maxIndex := 0

	var wg sync.WaitGroup
	var hashList sync.Map
	var uniqueHash sync.Map
	var lock sync.Mutex

	wg.Add(env.NUM_UPLOAD_WORKERS)

	for x := 0; x < env.NUM_UPLOAD_WORKERS; x++ {
		go func() {
			for x := range fileBlockChannel {
				lock.Lock()
				if x.Index > maxIndex {
					maxIndex = x.Index
				}
				lock.Unlock()
				hashList.Store(x.Index, x.Hash)
				_, ok := uniqueHash.LoadOrStore(x.Hash, true)
				if !ok {
					if checkForHashOnServer(x.Hash) {
						sendFileBlock(x.Hash, x.Bytes)
					}
				}
				bar.Add(len(x.Bytes))
			}
			wg.Done()
		}()
	}

	wg.Wait()
	bar.FinishPrint("Upload Complete")

	// we assemble an ordered hashlist of the file blocks
	finalList := make([]string, maxIndex+1)

	// iterate over the sent hashes, which include an index of where they are in the file
	// the index becomes the position in the final array.
	hashList.Range(func(key, value interface{}) bool {
		finalList[key.(int)] = value.(string)
		return true
	})

	// in order to send the list, we encode the slice to a byte format.
	var listStore bytes.Buffer
	enc := gob.NewEncoder(&listStore)
	enc.Encode(finalList)

	// and we send
	if sendHashList(fileName, &listStore) {
		fmt.Println("File Stored")
	}
}

func download(fileName string) {
	// recreate the file for a test to ./rebuilt.
	hashList := getHashList(fileName)

	fmt.Println("Workers On Download Pipeline:", env.NUM_DOWNLOAD_WORKERS)

	// add some emotion!
	bar := pb.StartNew(len(*hashList) * env.BLOCK * 1000).SetUnits(pb.U_BYTES)

	// build the workers
	workList := make(chan string, env.NUM_DOWNLOAD_WORKERS)
	var wg sync.WaitGroup
	for x := 0; x < env.NUM_DOWNLOAD_WORKERS; x++ {
		wg.Add(1)
		go workDownloads(workList, &wg, bar)
	}

	// do the work
	for _, x := range *hashList {
		wantFile := filepath.Join(env.DATASTORE, x)
		// check if we already have the file locally
		if !persist.FileExists(wantFile) {
			workList <- x
		} else {
			//file found locally
			bar.Add(env.BLOCK * 1000)
		}
	}

	// Clean up
	close(workList)
	wg.Wait()
	bar.FinishPrint("Download Complete, rebuilding...")
	hashing.Rebuild(hashList, env.DATASTORE, fileName+".rebuilt")
}

func deleteFile(fileName string) {
	ok := deleteFileFromServer(fileName)
	if ok {
		fmt.Println("Successfully Removed", fileName)
	} else {
		fmt.Println("Unable to remove", fileName)
	}
}

func list() {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/fileList")
	logging.PanicOnError("list request error", err)
	fmt.Println(readResponse(response))
}

func main() {
	env.BuildEnv()
	fileName, op := input.CollectOptions()
	logging.Initialize()

	if op == 'u' {
		upload(fileName)

	} else if op == 'd' {
		download(fileName)
	} else if op == 'r' {
		deleteFile(fileName)
	} else if op == 'l' {
		list()
	}
}
