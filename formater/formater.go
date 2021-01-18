package formater

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	recordMap["totalValidEmails"] = len(emails)
	categories := make(map[string]int)

	for _, email := range emails {
		if _, ok := categories[email]; ok {
			categories[email]++
		} else {
			categories[email] = 1
		}
	}
	recordMap["categories"] = categories
	validDomains := []string{}
	for k := range categories {
		validDomains = append(validDomains, k)
	}
	recordMap["valid-domains"] = validDomains
	return recordMap
}

func parseCSV(readers [][]string) map[string]interface{} {
	recordMap := make(map[string]interface{})
	url := "https://cloudflare-dns.com/dns-query?name=%s&type=AAAA"
	validEmails := regexp.MustCompile(`[a-zA-Z0-9]+@[a-zA-Z0-9\.]+\.[a-zA-Z0-9]+`)
	csvData := [][]string{{"Emails"}}
	// client := http.Client{
	// 	Timeout: time.Duration(15 * time.Second),
	// }

	for _, reader := range readers {
		if validEmails.MatchString(reader[0]) {
			// request, err := http.NewRequest("GET", fmt.Sprintf(url, strings.Split(reader[0], "@")[1]), nil)
			resp, err := http.Get(fmt.Sprintf(url, strings.Split(reader[0], "@")[1]))
			if err != nil {
				log.Println(err)
			}
			// request.Header.Set("Content-type", "application/dns-json")
			// resp, err := client.Do(request)
			// if err != nil {
			// 	log.Println(err)
			// }
			// var result map[string]interface{}
			// json.NewDecoder(resp.Body).Decode(&result)
			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
			}
			// log.Println(string(body))
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
	go processFile(cmdInput, validEmailStreams)
	for validEmail := range validEmailStreams {
		writeFile(filePath, validEmail, extended)
	}
}
