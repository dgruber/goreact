package main

import (
	"fmt"
	"os"

	"github.com/dgruber/goreact"
	calculator "github.com/mnogu/go-calculator"
)

func main() {

	openaiProvider, err := goreact.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		fmt.Printf("Failed to create OpenAIProvider: %v\n", err)
		os.Exit(1)
	}

	commands := map[string]goreact.Command{
		"calculate": {
			Name:        "calculate",
			Argument:    "expression",
			Description: "Calculate the answer to a math problem. The expression is like: 180*atan2(log(e), log10(10))/pi",
			Func: func(question string) (string, error) {
				result, err := calculator.Calculate(question)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("%f", result), nil
			},
		},
	}

	reactor, err := goreact.NewReact(openaiProvider, commands)
	if err != nil {
		fmt.Printf("Failed to create Reactor: %v\n", err)
		os.Exit(1)
	}

	answer, err := reactor.Question("What is the square root of 10? What is PI? What is the sum of both numbers?")
	if err != nil {
		fmt.Printf("Failed to get answer: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Answer: %s\n", answer)
}
