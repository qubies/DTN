package input

import (
	"github.com/pborman/getopt"
	"os"
)

var FILENAME string
var OPERATION byte
var HOH string

func CollectOptions() (string, byte) {
	filename_u := getopt.StringLong("upload", 'u', "", "The file you wish to upload")
	filename_d := getopt.StringLong("download", 'd', "", "The file you wish to download")
	filename_r := getopt.StringLong("remove", 'r', "", "The file you wish to remove")
	hoh_tmp := getopt.StringLong("HASH Value", 'h', "", "The specific hash value you wish to target (Remove and Download)")
	list := getopt.BoolLong("list", 'l', "Get a list of the files on the server by name")

	optHelp := getopt.BoolLong("help", 0, "Help")
	getopt.Parse()
	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}
	if *hoh_tmp != "" {
		HOH = *hoh_tmp
	}
	if *filename_u != "" {
		OPERATION = 'u'
		FILENAME = *filename_u
	} else if *filename_d != "" {
		OPERATION = 'd'
		FILENAME = *filename_d
	} else if *filename_r != "" {
		OPERATION = 'r'
		FILENAME = *filename_r

	} else if *list {
		OPERATION = 'l'
	} else {
		getopt.Usage()
	}
	return FILENAME, OPERATION
}
