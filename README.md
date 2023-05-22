# ReAct for Go

This is an implementation of ReAct for Go for fun. It combines Go functions with LLMs to let the LLM drive any Go code. The OpenAI
provider connects it to OpenAI's GPT.

## How it works

There is an initial question followed by an automatic thought, action and oberservation loop which runs until the LLM concludes that it has an answer.

The original paper describing the idea is [here](https://arxiv.org/pdf/2210.03629.pdf).

Here is a [a great blog post](https://blog.gopenai.com/react-a-bridge-between-llms-and-code-functions-54e5448c9a2) about the implementation in Python.

## How to use goreact?

First you need an instance of the LLM provider (OpenAI). Get the API key
from the OpenAI account and set the OPENAI_API_KEY env variable.

```go
	openaiProvider, err := goreact.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		fmt.Printf("Failed to create OpenAIProvider: %v\n", err)
		os.Exit(1)
	}
```

Then you need to register your actions which the LLM can work with. It requires
a function name, argument name, a description about how to use the function.
It should contain the function name and how the argument must look like and
of course what the output of the function is. Finally you need to declare the
function with string as input and output. The description, name, and argument
of the function is used in the prompt to tell the LLM about its existence.

````go
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
				content, err := page.GetContent()
				if err != nil {
					return err.Error(), err
				}
				return content, nil
			},
		},
	}
````

Then the agent can be used:

````go
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
````

## Examples

Examples from the examples directory.

### Calculator

Steering a calculator function.

```
QUESTION: What is the square root of 10? What is PI? What is the sum of both numbers?
THOUGHT: I need to calculate the square root of 10 and find the value of PI.
ACTION: { "command": "calculate", "args": "sqrt(10)" }
OBSERVATION: 3.162278
THOUGHT: I have the square root of 10, now I need to find the value of PI.
ACTION: { "command": "calculate", "args": "pi" }
OBSERVATION: 3.141593
THOUGHT: I have the value of PI, now I need to calculate the sum of both numbers.
ACTION: { "command": "calculate", "args": "3.162278+3.141593" }
OBSERVATION: 6.303871
Answer: I have the sum of the square root of 10 and PI.
ANSWER: The sum of the square root of 10 and PI is 6.303871.
```

### Rooms

Not even a robot - maybe a camera looking in different directions.

```
QUESTION: How many coins are in the rooms?
THOUGHT: I need to check all the rooms to count the coins.
ACTION: { "command": "look", "args": "north" } 
EXECUTING COMMAND: look north
OBSERVATION:  There is nothing north in bedroom
EXECUTING COMMAND: look east
OBSERVATION:  There is nothing east in bathroom
EXECUTING COMMAND: look south
OBSERVATION:  You found a coin in living room
EXECUTING COMMAND: look west
OBSERVATION:  There is nothing west in kitchen
Answer: I have checked all the rooms and found the coins.
ANSWER: There is 1 coin in the rooms.
```

## Wikisearch

Let the LLM work with wikipedia for finding the answer. Since the articles
are too long to fit into the context, it automatically summerizes the articles.

```
QUESTION: What is the fastest supercomputer today?
THOUGHT: I need to find out the fastest supercomputer today.
ACTION: { "command": "wikisearch", "args": "fastest supercomputer" }
...
OBSERVATION: Nothing interesting found
THOUGHT: I need more information to find the fastest supercomputer.
ACTION: { "command": "wikisearch", "args": "fastest supercomputer in the world" }
...
OBSERVATION: Fugaku was the fastest supercomputer in the world in June 2020, but was superseded by Frontier in May 2022.Fugaku is the fastest supercomputer today, with a performance of 442 petaFLOPS on the TOP500 FP64 high-performance LINPACK benchmark after its November 2020 upgrade. It also leads in other rankings, including Graph500, HPL-AI, and HPCG benchmark.
THOUGHT: Answer: I have found the answer to the question.
ANSWER: The fastest supercomputer today is Frontier, which superseded Fugaku in May 2022.
```

Another try:

```
QUESTION: What is the question for which the answer is 42?
THOUGHT: I need to search in Wikipedia for the question related to the number 42. 
ACTION:  { "command": "wikisearch", "args": "42 number significance" } 
EXECUTING COMMAND: wikisearch 42 number significance
OBSERVATION:  Nothing interesting found
EXECUTING COMMAND: wikisearch answer 42 question
Summarize observation
Summarize observation
OBSERVATION:  The specific question for which the answer is 42 is never found in the text. It is mentioned as the Answer to the Ultimate Question of Life, The Universe, and Everything. The text also explains that author Douglas Adams chose the number 42 as a joke, and it is an "ordinary number" with no hidden meaning.
Answer: I have found the reference to the answer 42 in a text by Douglas Adams.
ANSWER: The question for which the answer is 42 is known as the "Ultimate Question of Life, The Universe, and Everything" according to the book "The Hitchhiker's Guide to the Galaxy" by Douglas Adams. The specific question is never mentioned, and the number 42 is chosen as a joke with no hidden meaning.
```
