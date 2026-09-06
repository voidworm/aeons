package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"cli/cli"
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

		callReference, err := generateCommandReferenceFromInput(text)
		if err != nil {
			fmt.Println(err)
			continue
		}

		handler, err := getResourceHandler(callReference.Resource)
		if err != nil {
			fmt.Println(err)
			continue
		}
		handler(callReference)
	}
}

func getResourceHandler(resource string) (func(r cli.CallReference), error) {
	resourceToHandler := map[string]func(f cli.CallReference){
		"gateway":  handleGateway,
		"monolith": handleMonolith,
		"nebula":   handleNebula,
	}

	handler, ok := resourceToHandler[resource]
	if ok {
		return handler, nil
	} else {
		return nil, fmt.Errorf("received unknown resource type %s", resource)
	}
}

func handleGateway(input cli.CallReference) {
	fmt.Printf("handling gateway with verb %s and args %v\n", input.Verb, input.Params)
}
func handleMonolith(input cli.CallReference) {
	fmt.Printf("handling monolith with verb %s and args %v\n", input.Verb, input.Params)
}
func handleNebula(input cli.CallReference) {
	fmt.Printf("handling nebula with verb %s and args %s\v", input.Verb, input.Params)
}

func generateCommandReferenceFromInput(input string) (cli.CallReference, error) {

	cliName := ""
	resource := ""
	verb := ""
	args := ""
	argsMap := make(map[string]string)

	split := strings.Fields(input)

	if len(split) == 0 {
		return cli.CallReference{}, fmt.Errorf("no input provided")
	} else {
		//check if the leading command is actually the aeons cli
		if split[0] != "aeons" {
			return cli.CallReference{}, fmt.Errorf("the aeons have not been awakened")
		}
		cliName = split[0]
	}

	//catch if resource and verb were provided
	if len(split) == 1 {
		return cli.CallReference{}, fmt.Errorf("no resource provided")
	} else if len(split) == 2 {
		return cli.CallReference{}, fmt.Errorf("no verb provided")
	}

	resource = split[1]
	verb = split[2]

	//check if a param set was provided or not
	if len(split) > 3 {
		args = strings.Join(split[3:], "")
		argsArray := strings.Split(args, "-")
		//check if the params set actually starts on a -
		if strings.HasPrefix(argsArray[0], "-") {
			return cli.CallReference{}, fmt.Errorf("found input %s after verb, expected params declaration", split[3])
		}
		var err error


		argsMap, err = generateArgsMapFromArgs(argsArray)
		if (err != nil){
			return cli.CallReference{}, err
		}
	}

	inputAsCallReference := cli.CallReference{
		Name: cliName, 
		Resource: resource, 
		Verb: verb, 
		Params: argsMap,
	}
	return inputAsCallReference, nil
}

func generateArgsMapFromArgs(input []string) (map[string]string, error) {


	//this is huge mess.
	//i'll leave this here for now, to have a functioning reference for now that i can easily turn back on if I want to

	

	argsMap := make(map[string]string)

		for _, v := range input {
			fmt.Println(v)
		}

	for _, v := range input {
		argSplit := strings.Split(v, " ")
		if len(argSplit) > 2 {
			return nil, fmt.Errorf("arg %s contains multiple values %v", argSplit[0], argSplit[1:])
		}

		if len(argSplit) == 1 {
			argsMap[argSplit[0]] = ""
		} else {
			argsMap[argSplit[0]] = argSplit[1]
		}
	}
	return argsMap, nil
}
