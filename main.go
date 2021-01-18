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
	flag.Parse()
	if checkExtended(flag.Arg(1)) {
		*extended = true
	}
	if ok := getFileData(); ok {
		formater.Run(flag.Arg(0), *extended)
	}
}
