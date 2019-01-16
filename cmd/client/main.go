package main

import (
	"bytes"
	"fmt"
	env "github.com/qubies/DTN/env"
	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	"mime/multipart"
	"net/http"
)

func send(hash string, data []byte) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", hash)
	logging.PanicOnError("Error Creating Multipart Writer", err)
	part.Write(data)

	// I think this can be removed later.
	writer.WriteField("name", hash)

	err = writer.Close()
	logging.PanicOnError("Error Closing Multipart Writer", err)

	req, err := http.NewRequest("POST", "http://Localhost:"+env.RESTPORT+"/deposit", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	logging.PanicOnError("Error creating HTTP request", err)
	client := &http.Client{}
	resp, err := client.Do(req)
	logging.PanicOnError("Error sending HTTP request", err)
	var bodyContent []byte
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header)
	resp.Body.Read(bodyContent)
	resp.Body.Close()
	fmt.Println(bodyContent)
}

func main() {
	env.BuildEnv()
	logging.Initialize()

	// curently this just generates a hashlist for testing purposes.
	hl := new([]string)
	hashes, fileBlock := hash.GenerateHashList("testfile")
	// hashes, _ := hash.GenerateHashList("testfile")

	//build the persistent read write channels.
	hashStore := persist.NewFOB(env.HASHLIST, hl)
	hashStore.Object = hashes

	// persistently write and ensure file is on drive
	hashStore.WriteBlocking()

	test := persist.NewFOB(env.HASHLIST, hl)
	test.ReadBlocking()
	hashList := test.Object.(*[]string)
	for _, hash := range *hashList {
		fmt.Println("Hash:", hash)
	}
	for k, v := range fileBlock {
		send(k, v)
	}
}
