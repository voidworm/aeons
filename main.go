package main

import (
    "fmt"
    "os"
	"strings"
	"cli/cli"
)

func main() {

    args := os.Args[1:]

	if len(args) < 3{
		fmt.Println(fmt.Errorf("why u so bad at counting m8"))
		return;
	}

	paramsBaseArray := args[2:]
	params := map[string]string{}

	if len(paramsBaseArray) != 0 {
		err := fmt.Errorf("")
		params, err = generateArgsMapFromArgs(paramsBaseArray)
		if (err != nil){
			fmt.Println(err)
		}
	}

	ref := cli.CallReference{
		Resource: args[0],
		Verb: args[1],
		Params: params,
	}

	handler, err := getResourceHandler(ref.Resource)
	if err != nil {
		fmt.Println(err)
		return
	}
	handler(ref)
}

func getResourceHandler(resource string) (func(r cli.CallReference), error) {
	resourceToHandler := map[string]func(f cli.CallReference){
		"gateway":  handleGateway,
		"monolith": handleMonolith,
		"nebula":   handleNebula,
	}

	handler, ok := resourceToHandler[resource]
	if ok {
		return handler, nil;
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

func generateArgsMapFromArgs(input []string) (map[string]string, error){
	
	argsMap := make(map[string]string)
	for i := 0; i < len(input); i++ {
		
		currentArgument := input[i]

		if !strings.HasPrefix(currentArgument,"-") {
			return nil, fmt.Errorf("expected arg starting on - or flag starting on --, got %s",currentArgument)
		} else if strings.HasPrefix(currentArgument, "--"){
			//found a flag
			argsMap[currentArgument] = "YES"
		} else if strings.HasPrefix(currentArgument, "-"){
			//found and arg with value
			if (i == len(input)-1) || strings.HasPrefix(input[i+1],"-"){
				//found a flag with no value. ignore empty flag.
				fmt.Printf("found parameter %s without value, skipping %s\n", currentArgument, currentArgument)
			}else {
				argsMap[currentArgument] =input[i+1]
				i++ //the next item in the list is the value for this argument, skip evaluating it.
			}
		}
	}

	
	return argsMap, nil
}