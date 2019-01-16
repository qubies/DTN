package main

import (
	// "bytes"
	"fmt"
	env "github.com/qubies/DTN/env"
	hash "github.com/qubies/DTN/hashing"
	logging "github.com/qubies/DTN/logging"
	persist "github.com/qubies/DTN/persistentStore"
	// "mime/multipart"
	// "net/http"
)

// func send(hash [32]byte, data []byte) {
//     body := new(bytes.Buffer)
//     writer := multipart.NewWriter(body)
//     part, err := writer.CreateFormFile(paramName, fi.Name())
//     if err != nil {
//         return nil, err
//     }
//     part.Write()

//     for key, val := range params {
//         _ = writer.WriteField(key, val)
//     }
//     err = writer.Close()
//     if err != nil {
//         return nil, err
//     }

//     return http.NewRequest("POST", uri, body)
// }

func main() {
	env.BuildEnv()
	logging.Initialize()

	// curently this just generates a hashlist for testing purposes.
	hl := new([]string)
	// hashes, fileBlock := hash.GenerateHashList("testfile")
	hashes, _ := hash.GenerateHashList("testfile")

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
}
