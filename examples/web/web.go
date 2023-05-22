package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/dgruber/goreact"
	googlesearch "github.com/rocketlaunchr/google-search"

	"jaytaylor.com/html2text"
)

func main() {

	openaiProvider, err := goreact.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		fmt.Printf("Failed to create OpenAIProvider: %v\n", err)
		os.Exit(1)
	}

	openaiProvider.WithModel("gpt-3.5-turbo-0301")

	commands := map[string]goreact.Command{
		"scrape": {
			Name:        "scrape",
			Argument:    "http address",
			Description: "Scrape reads the content of a web page given by the http address",
			Func: func(address string) (string, error) {
				// get the content of the web page
				resp, err := http.Get(address)
				if err != nil {
					fmt.Printf("Failed to get web page %s: %v", address, err)
					return err.Error(), nil
				}
				defer resp.Body.Close()

				// convert HTML to text
				text, err := html2text.FromReader(resp.Body, html2text.Options{TextOnly: true})
				if err != nil {
					fmt.Printf("Failed to convert HTML to text: %v", err)
					return "", err
				}
				fmt.Printf("Scraped text with length %d\n", len(text))
				// write text to file in the current directory
				ioutil.WriteFile(fmt.Sprintf("./scrape-%s-%d.txt",
					address, time.Now().Unix()), []byte(text), 0644)

				return text, nil
			},
		},
		"search": {
			Name:        "search",
			Argument:    "search term",
			Description: "Search for a term on Google",
			Func: func(term string) (string, error) {
				var opts = googlesearch.SearchOptions{
					CountryCode:    "de",
					LanguageCode:   "en",
					Limit:          5,
					Start:          0,
					OverLimit:      false,
					FollowNextPage: false,
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"}
				result, err := googlesearch.Search(context.Background(), term, opts)
				if err != nil {
					fmt.Printf("Search failed with error: %v\n", err)
					return "search failed with error: " + err.Error(), err
				}
				fmt.Printf("Search result: %v\n", result)
				return fmt.Sprintf("%v", result), nil
			},
		},
		"ask": {
			Name:        "ask",
			Argument:    "question",
			Description: "Ask a question to the user for further clarification of the question",
			Func: func(term string) (string, error) {
				// get interactive answer from user
				fmt.Printf("Please answer the question: %s\n", term)
				reader := bufio.NewReader(os.Stdin)
				text, _ := reader.ReadString('\n')
				return text, nil
			},
		},
	}

	reactor, err := goreact.NewReact(openaiProvider, commands)
	if err != nil {
		fmt.Printf("Failed to create Reactor: %v\n", err)
		os.Exit(1)
	}

	answer, err := reactor.Question("What is the answer to life, the universe and everything?")
	if err != nil {
		fmt.Printf("Failed to get answer: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CONCLUSION: %s\n", answer)
}
