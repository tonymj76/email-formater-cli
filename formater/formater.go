package formater

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type inputFile struct {
	filePath string
	extended bool
}

func newInputFile(filePath string, extended bool) *inputFile {
	return &inputFile{
		filePath: filePath,
		extended: extended,
	}
}

func checkIfValidFile(filename string) (bool, error) {
	if fileExtendsion := filepath.Ext(filename); fileExtendsion != ".csv" {
		return false, fmt.Errorf("file %s is not CSV", filename)
	}
	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return false, fmt.Errorf("file %s does not exist", filename)
	}
	return true, nil
}

func parseJSON(readers [][]string) map[string]interface{} {
	recordMap := make(map[string]interface{})
	recordMap["totalEmailsParsed"] = len(readers)
	validEmails := regexp.MustCompile(`[a-zA-Z0-9]+@[a-zA-Z0-9\.]+\.[a-zA-Z0-9]+`)
	emails := []string{}
	for _, reader := range readers {
		if validEmails.MatchString(reader[0]) {
			emails = append(emails, strings.Split(reader[0], "@")[1])
		}
	}
	recordMap["valid-domains"] = emails
	categories := make(map[string]int)

	for _, email := range emails {
		if _, ok := categories[email]; ok {
			categories[email]++
		} else {
			categories[email] = 1
		}
	}
	recordMap["categories"] = categories
	return recordMap
}

func parseCSV(readers [][]string) map[string]interface{} {
	recordMap := make(map[string]interface{})
	validEmails := regexp.MustCompile(`[a-zA-Z0-9]+@[a-zA-Z0-9\.]+\.[a-zA-Z0-9]+`)
	csvData := [][]string{{"Emails"}}
	for _, reader := range readers {
		if validEmails.MatchString(reader[0]) {
			csvData = append(csvData, reader)
		}
	}
	recordMap["mails"] = csvData
	return recordMap
}

func processFile(fileData *inputFile, writerStream chan<- map[string]interface{}) {
	file, err := os.Open(fileData.filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)

	_, err = reader.Read() // taking the headers out
	if err != nil {
		log.Fatal(err)
	}
	readersData, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	if len(readersData) < 1 {
		log.Fatal("no data to read")
	}
	defer close(writerStream)
	formatedEmails := make(map[string]interface{})
	if fileData.extended {
		formatedEmails = parseCSV(readersData)
	} else {
		formatedEmails = parseJSON(readersData)
	}
	writerStream <- formatedEmails
}

func writeFile(filePath string, fileType map[string]interface{}, extended bool) {
	filename := strings.Split(filepath.Base(filePath), ".")[0]
	if extended {
		file, err := os.OpenFile(fmt.Sprintf("%s.csv", filename), os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		csvfile := csv.NewWriter(file)
		csvfile.WriteAll(fileType["mails"].([][]string))
		if err := csvfile.Error(); err != nil {
			log.Fatalln("error writing csv:", err)
		}
	} else {
		file, err := os.OpenFile(fmt.Sprintf("%s.json", filename), os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		enc.Encode(fileType)
	}
}

//Run power house
func Run(filePath string, extended bool) {
	validEmailStreams := make(chan map[string]interface{})
	ok, err := checkIfValidFile(filePath)
	if !ok {
		log.Fatal(err)
	}
	cmdInput := newInputFile(filePath, extended)
	processFile(cmdInput, validEmailStreams)
	for validEmail := range validEmailStreams {
		writeFile(filePath, validEmail, extended)
	}
}
