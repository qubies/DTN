package main

import (
	"bytes"
	"fmt"
	env "github.com/qubies/DTN/env"
	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

// TODO this is really inefficient as it requires the entire file to be loaded in memory
// It is likey that this shoul dbe a worker on the channel producing the files

func send(hash string, data []byte) {
	// adapted from https://gist.github.com/mattetti/5914158/f4d1393d83ebedc682a3c8e7bdc6b49670083b84
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", hash)
	logging.PanicOnError("Error Creating Multipart Writer", err)
	part.Write(data)

	// I think this can be removed later.
	writer.WriteField("name", hash)

	err = writer.Close()
	logging.PanicOnError("Error Closing Multipart Writer", err)

	req, err := http.NewRequest("POST", "http://"+env.SERVER_URL+":"+env.RESTPORT+"/deposit", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	logging.PanicOnError("Error creating HTTP request", err)
	client := &http.Client{}
	resp, err := client.Do(req)
	logging.PanicOnError("Error sending HTTP request", err)
	var bodyContent []byte
	// fmt.Println(resp.StatusCode)
	// fmt.Println(resp.Header)
	resp.Body.Read(bodyContent)
	resp.Body.Close()
}

func check(hash string) bool {
	response, err := http.Get("http://" + env.SERVER_URL + ":" + env.RESTPORT + "/check?hash=" + hash)
	logging.PanicOnError("Get Request to checker", err)
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	logging.PanicOnError("reading get request body from checker", err)
	fmt.Println("Received:", string(contents))
	return string(contents) == "SEND"
}

func main() {
	env.BuildEnv()
	logging.Initialize()

	// curently this just generates a hashlist for testing purposes.
	hl := new([]string)
	partChan := hash.GenerateHashList("testfile")
	var hashList []string
	for x := range partChan {
		hashList = append(hashList, x.Hash)
		if check(x.Hash) {
			send(x.Hash, x.Bytes)
			fmt.Println("Sent Hash:", x.Hash)
		} else {
			fmt.Println("Hash:", x.Hash)
		}
	}

	//build the persistent read write channels.
	hashStore := persist.NewFOB(env.HASHLIST, hl)
	hashStore.Object = hashList

	// persistently write and ensure file is on drive
	hashStore.WriteBlocking()

	test := persist.NewFOB(env.HASHLIST, hl)
	test.ReadBlocking()
}
