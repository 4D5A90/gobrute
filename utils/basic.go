package utils

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
)

func IsValidURL(u string) bool {
	parsedURL, err := url.Parse(u)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func IsFile(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}

func GetFlagAsList(f string) []string {

	if IsFile(f) {
		readFile, err := os.Open(f)

		if err != nil {
			fmt.Println(err)
		}
		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)
		var fileLines []string

		for fileScanner.Scan() {
			fileLines = append(fileLines, fileScanner.Text())
		}

		readFile.Close()

		return fileLines
	} else {
		return []string{f}
	}
}
