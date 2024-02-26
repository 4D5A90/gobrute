/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"gobrute/utils"
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

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json [target] [payload]",
	Short: "HTTP (Post) JSON brute-force attack",
	Long:  `A command-line tool that performs HTTP (Post) JSON brute-force attacks using a specified URL and payload.`,
	Run: func(cmd *cobra.Command, args []string) {
		// This function will be called when the root command is executed
		if len(args) < 2 {
			fmt.Println("Please provide a URL (target) and a payload")
			return
		}

		target = args[0]

		if !utils.IsValidURL(target) {
			fmt.Println("Please provide a valid URL (target)")
			return
		}

		payload = args[1]

		crack(target, payload, utils.GetFlagAsList(user), utils.GetFlagAsList(pass))
	},
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

func init() {
	rootCmd.AddCommand(jsonCmd)
	jsonCmd.Flags().StringVar(&user, "user", "", "Single Username or Email or Path to the user file")
	jsonCmd.Flags().StringVar(&pass, "pass", "", "Single password or Path to the password file")
	jsonCmd.Flags().StringVar(&payload, "payload", "", "JSON payload")
}
