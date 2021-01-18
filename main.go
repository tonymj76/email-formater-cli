package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/tonymj76/email-formater-cli/formater"
)

func getFileData() bool {
	if len(os.Args) == 1 {
		fmt.Println("Timothy Stephen Mayor")
		return false
	}
	return true
}
func checkExtended(str string) bool {
	validFlag := regexp.MustCompile("extended")
	return validFlag.MatchString(str)
}

func main() {
	extended := flag.Bool("extended", false, "check MX record")
	help := flag.Bool("help", false, "help on commands")
	flag.Parse()
	if *help {
		fmt.Println(`
		timothy-stephen v1 of January 2021 by Timothy Stephen Mayor
		usage: timothy-stephen <file input must be CSV> [flag]
		
		flags:
			--help
			--extended

		example:

		timothy-stephen --help display help information
		timothy-stephen <file input must be CSV> this prints a json formatter at the root of the directory
		timothy-stephtn <file input> --extended 
		`)
		return
	}
	if checkExtended(flag.Arg(1)) {
		*extended = true
	}
	if ok := getFileData(); ok {
		formater.Run(flag.Arg(0), *extended)
	}
}
