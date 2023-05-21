package main

import (
	"fmt"
	"os"

	"github.com/dgruber/goreact"
)

func main() {

	openaiProvider, err := goreact.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		fmt.Printf("Failed to create OpenAIProvider: %v\n", err)
		os.Exit(1)
	}

	type Room struct {
		name      string
		hasCoin   bool
		neighbors map[string]Room
	}

	floorPlan := Room{
		"hallway",
		false,
		map[string]Room{
			"west": {
				"kitchen",
				false,
				nil,
			},
			"north": {
				"bedroom",
				false,
				nil,
			},
			"east": {
				"bathroom",
				false,
				nil,
			},
			"south": {
				"living room",
				true,
				nil,
			},
		},
	}

	commands := map[string]goreact.Command{
		"look": {
			Name:        "look",
			Argument:    "direction",
			Description: "Looks into a room which is north, south, east, or west. The result is a description of the room and tells if it contains a coin.",
			Func: func(direction string) (string, error) {
				room, ok := floorPlan.neighbors[direction]
				if !ok {
					return "You see a wall.", nil
				}
				if room.hasCoin {
					return "You found a coin in " + room.name, nil
				}
				return "There is nothing " + direction + " in " + room.name, nil
			},
		},
	}

	reactor, err := goreact.NewReact(openaiProvider, commands)
	if err != nil {
		fmt.Printf("Failed to create Reactor: %v\n", err)
		os.Exit(1)
	}

	answer, err := reactor.Question("How many coins are in the rooms?")
	if err != nil {
		fmt.Printf("Failed to get answer: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Answer: %s\n", answer)
}
