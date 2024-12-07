package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func parseCommand(input string) (string, []string) {
	var cmdName string
	var args []string

	fmt.Println("Input:", input) // Debug: Print the input

	// Check for single quotes
	if strings.Contains(input, "'") {
		re := regexp.MustCompile("'(.*?)'")
		matches := re.FindAllStringSubmatch(input, -1)

		for _, match := range matches {
			if len(match) > 1 {
				args = append(args, match[1]) // Add the content inside single quotes
			}
		}
	} else {
		// Handle normal input without single quotes
		args = strings.Fields(input)
	}

	if len(args) > 0 {
		cmdName = args[0] // The first element is the command name
		args = args[1:]   // The rest are arguments
	}

	fmt.Printf("Command: '%s', Arguments: %v\n", cmdName, args) // Debug: Print command and arguments

	return cmdName, args
}

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		s, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		s = strings.Trim(s, "\r\n")
		
		cmd, args := parseCommand(s)
		
		switch cmd {
		case "echo":
			fmt.Println(strings.Join(args, " "))
			continue
		case "exit":
			os.Exit(0)
			continue
		default:
			fmt.Println("Command not recognized.")
			continue
		}
	}
}