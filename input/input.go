package input

import (
	"fmt"
	"github.com/pborman/getopt"
	"os"
)

func GetFile() string {
	// modelled from https://stackoverflow.com/questions/1714236/getopt-like-behavior-in-go
	uploadName := getopt.StringLong("upload", 'u', "", "The file you wish to upload")
	optHelp := getopt.BoolLong("help", 0, "Help")
	getopt.Parse()
	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}
	fmt.Println("Working on File:", *uploadName)

	return *uploadName

}
