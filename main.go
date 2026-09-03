package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Println("Echo of Aeons")
	fmt.Println("---------------------")
	for {
		fmt.Print("<<< ")
		text, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(">>> ")

		resource, verb, args, err := splitInput(text)
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch resource {
		case "gateway":
			handleGateway(verb, args)
		case "monolith":
			handleMonolith(verb, args)
		case "nebula":
			handleNebula(verb, args)
		default:
			fmt.Printf("aeons resource type issue: received unknown resource %s", resource)

		}
	}
}

func handleGateway(verb string, args string) {
	fmt.Printf("handling gateway with verb %s and args %s\n", verb, args)
}
func handleMonolith(verb string, args string) {
	fmt.Printf("handling monolith with verb %s and args %s\n", verb, args)
}
func handleNebula(verb string, args string) {
	fmt.Printf("handling nebula with verb %s and args %s\n", verb, args)
}

func splitInput(input string) (string, string, string, error) {

	resource := ""
	verb := ""
	args := ""

	split := strings.Fields(input)

	if len(split) == 0 {
		return "", "", "", fmt.Errorf("no input provided")
	} else {
		//check if the leading command is actually the aeons cli
		if split[0] != "aeons" {
			return "", "", "", fmt.Errorf("the aeons have not been awakened")
		}
	}

	//catch if resource and verb were provided
	if len(split) == 1 {
		return "", "", "", fmt.Errorf("no resource provided")
	} else if len(split) == 2 {
		return "", "", "", fmt.Errorf("no verb provided")
	}

	resource = split[1]
	verb = split[2]

	//check if a param set was provided or not
	if len(split) > 3 {
		args = strings.Join(split[3:], " ")
		//check if the params set actually starts on a -
		if !(strings.HasPrefix(args, "-")) {
			return "", "", "", fmt.Errorf("found input %s after verb, expected params declaration", split[3])
		}
	}

	return resource, verb, args, nil
}
