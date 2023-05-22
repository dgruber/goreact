package goreact

import (
	"fmt"
	"strings"
)

type Command struct {
	Name        string
	Argument    string
	Description string
	Func        func(string) (string, error)
}

type React struct {
	llm      LLMProvider
	commands map[string]Command
}

func NewReact(llmProvider LLMProvider, commands map[string]Command) (*React, error) {
	if commands == nil {
		return nil, fmt.Errorf("commands cannot be nil")
	}
	return &React{
		llm:      llmProvider,
		commands: commands,
	}, nil
}

func (r *React) Question(question string) (string, error) {
	fmt.Println("QUESTION:", question)
	fullprompt, action, answer, err := r.getInitialThoughtAndAction(question)
	if err != nil {
		return "", err
	}
	if answer != "" {
		// Looks like the LLM very often answers the question directly.
		// We don't really want that, it should use the commands at least
		// once. Hence I added an instruction in the prompt to run
		// at least one cycle...
		fmt.Printf("Too easy. Immediately answering: %s\n", answer)
		return answer, nil
	}

	fmt.Println(action)
	observation, err := r.executeAction(action)
	if err != nil {
		return "", err
	}

	// The observation of the action might be too long to serve
	// as input for the next step. Hence we compress it to 512
	// characters. Doing that by letting the LLM summarize the
	// observation based on relevant information with regards
	// to the question.
	observation, err = r.createSummaryOfSummaries(question, observation, 512)
	if err != nil {
		return "", fmt.Errorf("unable to compress observation: %v", err)
	}

	for {
		fmt.Println("OBSERVATION: ", observation)
		if strings.Contains(observation, "ANSWER:") {
			fmt.Println("ANSWER:", observation)
			return observation, nil
		}
		fullprompt = fullprompt + "\n" + "OBSERVATION: " + observation

		fullprompt, action, err = r.getThoughtAndAction(fullprompt)
		if err != nil {
			if strings.Contains(fullprompt, "ANSWER:") {
				return fullprompt, nil
			}
			return "", err
		}
		if strings.Contains(fullprompt, "ANSWER:") {
			return fullprompt, nil
		}

		observation, err = r.executeAction(action)
		if err != nil && observation == "" {
			// observation might contain the error of the application
			// which can be helpful to understand what went wrong for
			// the LLM. Hence we only return "hard" errors which has
			// no observation to abort the conversation.
			return "", err
		}

		observation, err = r.createSummaryOfSummaries(question, observation, 512)
		if err != nil {
			return "", fmt.Errorf("unable to compress observation: %v", err)
		}
	}
}

func (r *React) getInitialThoughtAndAction(question string) (string, string, string, error) {
	system := fmt.Sprintf(BasicReActPrompt, r.commandDescriptions())

	fullprompt, err := r.llm.Request(system, "QUESTION: "+question+"\n")
	if err != nil {
		return "", "", "", err
	}
	if strings.Contains(fullprompt, "ANSWER: ") {
		return "", "", strings.Split(fullprompt, "ANSWER:")[1], nil
	}

	fullprompt = "QUESTION: " + question + "\n" + fullprompt

	lines := strings.Split(fullprompt, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "THOUGHT: ") {
			fmt.Println(line)
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "ACTION: ") {
			fmt.Println(line)
			// parse ACTION: from result
			result := strings.Split(fullprompt, "ACTION: ")
			if len(result) != 2 {
				return "", "", "", fmt.Errorf("unable to parse initial ACTION: from result: %s", fullprompt)
			}
			return fullprompt, result[1], "", nil
		}
	}
	return "", "", "", fmt.Errorf("unable to parse initial ACTION: from result: %s", fullprompt)
}

func (r *React) getThoughtAndAction(history string) (string, string, error) {
	prompt := fmt.Sprintf("%s\nTHOUGHT: ", history)
	system := fmt.Sprintf(BasicReActPrompt, r.commandDescriptions())

	thought, err := r.llm.Request(system, prompt)
	if err != nil {
		return "", "", err
	}

	thought = strings.Trim(thought, "\n")

	// check if there is an answer
	if strings.Contains(thought, "ANSWER: ") {
		return thought, "", nil
	}

	fmt.Println("THOUGHT: " + strings.Split(thought, "\n")[0])

	// parse ACTION: from result
	result := strings.Split(thought, "ACTION: ")
	if len(result) != 2 {
		return "", "", fmt.Errorf("unable to parse ACTION: from result: %s", thought)
	}
	return fmt.Sprintf("%s\nTHOUGHT: %s", history, thought), result[1], nil
}

func (r *React) commandDescriptions() string {
	var descriptions []string
	descriptions = append(descriptions, "")
	descriptions = append(descriptions, "command | argument | description")
	descriptions = append(descriptions, "--------------------------------")
	for name, command := range r.commands {
		descriptions = append(descriptions, fmt.Sprintf("%s | %s | %s",
			name, command.Argument, command.Description))
	}
	descriptions = append(descriptions, "--------------------------------")
	descriptions = append(descriptions, "")
	return strings.Join(descriptions, "\n")
}

func (r *React) executeAction(action string) (string, error) {
	command, argument, err := parseAction(action)
	if err != nil {
		return "", err
	}
	cmd, exists := r.commands[command]
	if !exists {
		return "", fmt.Errorf("unknown command: %s", command)
	}
	fmt.Printf("EXECUTING COMMAND: %s %s\n", command, argument)
	return cmd.Func(argument)
}

func (r *React) createSummaryOfSummaries(question, observation string, maxLen int) (string, error) {
	var err error
	if len(observation) <= maxLen {
		return observation, nil
	}
	for {
		observation, err = r.compressObservation(question, observation, maxLen)
		if err != nil {
			return "", fmt.Errorf("unable to compress observation: %v", err)
		}
		if len(observation) > maxLen {
			// create a summary of the summary
			fmt.Printf("Summary too long (%d). Creating a summary of the summary.\n",
				len(observation))
			continue
		} else {
			break
		}
	}
	return observation, nil
}

func (r *React) compressObservation(question, observation string, maxLen int) (string, error) {
	// compress observation
	if len(observation) <= maxLen {
		return observation, nil
	}
	// go through the observation with a sliding window and
	// create a summary which is related to the question
	// take first 2048 characters (roughly 512 tokens)
	stepSize := 2048
	overlap := 32

	from := 0
	to := stepSize

	fullSummary := ""

	for {
		if to > len(observation) {
			to = len(observation) - 1
		}
		part := observation[from:to]

		summary, err := r.llm.Request(PromptSummarize,
			"Question: "+question+"\n"+"Here is the text to summarize:\n"+part+"\n")
		if err != nil {
			return "", fmt.Errorf("failed to summarize observation: %v", err)
		}
		// remove EMPTY
		summary = strings.ReplaceAll(summary, "EMPTY", "")
		fullSummary += summary
		from += stepSize - overlap
		if from > len(observation)-overlap {
			break
		}
		to += stepSize - overlap
	}
	// remove empty lines
	fullSummary = strings.Trim(fullSummary, "\n")
	if len(fullSummary) == 0 {
		return "Nothing interesting found", nil
	}
	return strings.Trim(fullSummary, "\n"), nil
}

func parseAction(action string) (string, string, error) {
	// action contains a command and an argument
	// { "command": "answer", "argument": "42" }
	result := strings.Split(strings.TrimSuffix(action, "\n"), "}")

	commandAndArgument := strings.Split(strings.Trim(result[0], " "), ", ")
	if len(commandAndArgument) != 2 {
		return "", "", fmt.Errorf("unable to parse action: %s", action)
	}

	command := strings.Split(commandAndArgument[0], ": ")
	if len(command) != 2 {
		return "", "", fmt.Errorf("unable to parse action: %s", action)
	}
	cmd := strings.Trim(command[1], "\" ")

	argument := strings.Split(commandAndArgument[1], ": ")
	if len(argument) != 2 {
		return "", "", fmt.Errorf("unable to parse action: %s", action)

	}
	arg := strings.Trim(argument[1], "\" ")

	return cmd, arg, nil
}
