package main

import (
	"fmt"
	"os"

	"github.com/dgruber/goreact"
	gowiki "github.com/trietmn/go-wiki"
)

func main() {

	openaiProvider, err := goreact.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		fmt.Printf("Failed to create OpenAIProvider: %v\n", err)
		os.Exit(1)
	}

	commands := map[string]goreact.Command{
		"wikisearch": {
			Name:        "wikisearch",
			Argument:    "topic",
			Description: "wikisearch searches Wikipedia for a topic",
			Func: func(topic string) (string, error) {
				page, err := gowiki.GetPage(topic, -1, false, true)
				if err != nil {
					return "Topic " + topic + "not found in Wikipedia", err
				}
				// Get the content of the page
				content, err := page.GetContent()
				if err != nil {
					return err.Error(), err
				}
				return content, nil
			},
		},
	}

	reactor, err := goreact.NewReact(openaiProvider, commands)
	if err != nil {
		fmt.Printf("Failed to create Reactor: %v\n", err)
		os.Exit(1)
	}

	answer, err := reactor.Question("What is the question for which the answer is 42?")
	if err != nil {
		fmt.Printf("Failed to get answer: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CONCLUSION: %s\n", answer)
}
