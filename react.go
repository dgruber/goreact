package goreact

import (
	"fmt"
	"strings"
	"time"
)

type Command struct {
	Name        string
	Argument    string
	Description string
	Func        func(string) (string, error)
}

type React struct {
	llm        LLMProvider
	commands   map[string]Command
	mainPrompt string
}

func NewReact(llmProvider LLMProvider, commands map[string]Command) (*React, error) {
	if commands == nil {
		return nil, fmt.Errorf("commands cannot be nil")
	}
	return &React{
		llm:        llmProvider,
		commands:   commands,
		mainPrompt: BasicReActPrompt,
	}, nil
}

func (r *React) WithMainPrompt(prompt string) *React {
	r.mainPrompt = prompt
	return r
}

func (r *React) Question(question string) (string, error) {
	fmt.Println("QUESTION:", question)
	fullprompt, action, answer, err := r.getInitialThoughtAndAction(r.mainPrompt, question)
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
		if strings.Contains(fullprompt, "ANSWER:") {
			return fullprompt, nil
		}
		if err != nil {
			return "", err
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

func (r *React) getInitialThoughtAndAction(systemPrompt, question string) (string, string, string, error) {
	systemPrompt = fmt.Sprintf(systemPrompt, r.commandDescriptions())
	fullPrompt, err := r.llm.Request(systemPrompt, "QUESTION: "+question+"\n")
	if err != nil {
		return "", "", "", err
	}

	if strings.Contains(fullPrompt, "ANSWER: ") {
		return "", "", strings.Split(fullPrompt, "ANSWER:")[1], nil
	}

	fullPrompt = "QUESTION: " + question + "\n" + fullPrompt

	for _, line := range strings.Split(fullPrompt, "\n") {
		if strings.HasPrefix(line, "THOUGHT: ") {
			fmt.Println(line)
		}
		if strings.HasPrefix(line, "ACTION: ") {
			fmt.Println(line)
			return fullPrompt, strings.Replace(line, "ACTION: ", "", 1), "", nil
		}
	}

	return "", "", "", fmt.Errorf("no action found: %s", fullPrompt)
}

func (r *React) getThoughtAndAction(history string) (string, string, error) {
	prompt := fmt.Sprintf("%s\nTHOUGHT: ", history)
	system := fmt.Sprintf(BasicReActPrompt, r.commandDescriptions())

	fmt.Printf("[context size: around %d tokens]\n", len(prompt)/4)
	if len(prompt)/4 > 14000 {
		fmt.Println("WARNING: context size is too large, truncating...")
		// remove all lines starting with OBSERVATION:
		prompt = compressPromptContext(prompt)
	}
	thought, err := r.llm.Request(system, prompt)
	if err != nil {
		return "", "", err
	}

	thought = strings.Trim(thought, "\n")

	// check if there is an answer
	if strings.Contains(thought, "ANSWER: ") {
		return thought, "", nil
	}

	// THOUGHTS can be multilines
	fmt.Println("THOUGHT: " + strings.Split(thought, "\n")[0])

	// parse ACTION: from result
	var result []string
	for {
		result = strings.Split(thought, "ACTION: ")
		if len(result) < 2 {
			// there is no ACTION: retry
			prompt := fmt.Sprintf("%s\nACTION: ", history+"\nnTHOUGHT: "+thought)
			thought, err = r.llm.Request(system, prompt)
			if err != nil {
				return "", "", err
			}
			thought = strings.Trim(thought, "\n")
		} else {
			break
		}
	}
	return fmt.Sprintf("%s\nTHOUGHT: %s", history, thought), result[len(result)-1], nil
}

func compressPromptContext(prompt string) string {
	// remove all lines starting with OBSERVATION:
	lines := strings.Split(prompt, "\n")
	var newlines []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "OBSERVATION: ") {
			newlines = append(newlines, line)
		}
	}
	prompt = strings.Join(newlines, "\n")
	fmt.Printf(prompt)
	fmt.Printf("[context size: around %d tokens]\n", len(prompt)/4)
	return prompt
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
	command, argument, err := parseAction2(action)
	if err != nil {
		return "", err
	}
	cmd, exists := r.commands[command]
	if !exists {
		return fmt.Sprintf("The command %s is not known. Please use one of the following commands:\n%s",
			command, r.commandDescriptions()), nil
		//return "", fmt.Errorf("unknown command: %s", command)
	}
	fmt.Printf("EXECUTING COMMAND: %s %s\n", command, argument)
	return cmd.Func(argument)
}

func (r *React) createSummaryOfSummaries(question, observation string, maxLen int) (string, error) {
	var err error
	if len(observation) <= maxLen {
		return observation, nil
	}
	// get last line which contains THOUGHT
	thought := ""
	for _, line := range strings.Split(observation, "\n") {
		if strings.Contains(line, "THOUGHT:") {
			thought = line
		}
	}

	for {

		before := len(observation)
		observation, err = r.compressObservation(question+" "+thought, observation, maxLen)
		if err != nil {
			return "", fmt.Errorf("unable to compress observation: %v", err)
		}
		after := len(observation)
		if before <= after {
			// it does not get shorter
			observation, err = r.llm.Request("Summarize in 3 sentences according to the question.",
				"Question: "+question+"\n"+"Here is the text to summarize in 3 sentences:\n"+observation+"\n")
			break
		}

		if len(observation) > maxLen {
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

		var summary string
		var err error

		for {
			summary, err = r.llm.Request(PromptSummarize,
				"Question: "+question+"\n"+"Here is the text to summarize in two sentences:\n"+part+"\n")
			if err != nil {
				if strings.Contains(err.Error(), "currently overloaded") {
					// retry since API is overloaded...
					<-time.Tick(5 * time.Second)
					continue
				}
				return "", fmt.Errorf("failed to summarize observation: %v", err)
			}
			break
		}

		summary = strings.ReplaceAll(summary, "EMPTY", "")
		fullSummary += summary
		from += stepSize - overlap
		if from > len(observation)-overlap {
			break
		}
		to += stepSize - overlap
	}

	fullSummary = strings.Trim(fullSummary, "\n")
	if len(fullSummary) == 0 {
		return "Nothing interesting found", nil
	}
	return fullSummary, nil
}

func parseAction2(action string) (string, string, error) {
	result := strings.Split(strings.TrimSuffix(action, "\n"), "\n")
	resultAndArgs := strings.Split(result[0], " ")
	if len(resultAndArgs) < 2 {
		return resultAndArgs[0], "", nil
	}
	return resultAndArgs[0], strings.Trim(strings.Join(resultAndArgs[1:], " "), " \""), nil
}

func parseAction(action string) (string, string, error) {
	// action contains a command and an argument
	// { "command": "answer", "argument": "42" }
	// answer: arguments for answer
	result := strings.Split(strings.TrimSuffix(action, "\n"), "}")

	commandAndArgument := strings.Split(strings.Trim(result[0], " "), ", ")
	if len(commandAndArgument) < 2 {
		return "", "", fmt.Errorf("unable to parse action: %s", action)
	}

	command := strings.Split(commandAndArgument[0], ": ")
	if len(command) < 2 {
		return "", "", fmt.Errorf("unable to parse action command: %s", action)
	}
	cmd := strings.Trim(command[1], "\" ")

	argument := strings.Split(strings.Join(commandAndArgument[1:], ", "), ": ")
	if len(argument) < 2 {
		return "", "", fmt.Errorf("unable to parse action argument: %s", action)

	}
	arg := strings.Trim(strings.Join(argument[1:], ":"), "\" ")

	return cmd, arg, nil
}
