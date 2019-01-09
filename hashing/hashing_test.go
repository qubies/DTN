package Hashing

import (
	"bytes"
	"testing"
)

const ITERATIONS = 10

func TestFileList(t *testing.T) {
	// first we generate an initial set of hashes.
	initial := GenerateHashList("testfile")
	for x := 0; x < ITERATIONS; x++ {
		new := GenerateHashList("testfile")
		if len(new) != len(initial) {
			t.Error("Hash lengths do not match\n")
		}
		for i, v := range initial {
			if !bytes.Equal(v[:], new[i][:]) {
				t.Error("Hashes do not match\n")
			}
		}
	}
	new := GenerateHashList("testfile2")
	for i, v := range new { // NOTE that testfiles2 must be smaller than testfile
		if bytes.Equal(v[:], initial[i][:]) {
			t.Error("Collision Detected??\n")
		}
	}
}
