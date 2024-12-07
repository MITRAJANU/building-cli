package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	for {
		// Prompt for user input
		fmt.Fprint(os.Stdout, "$ ")

		// Read user input
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Trim whitespace and newlines from input
		command := strings.TrimSpace(input)

		// Check if the command is empty
		if command == "" {
			continue // Ignore empty commands
		}

		// Split the command into name and arguments
		args := strings.Fields(command)
		cmdName := args[0]
		
		// Execute the command and handle errors
		if err := executeCommand(cmdName, args[1:]); err != nil {
			fmt.Printf("%s: %s\n", cmdName, err.Error())
		}
	}
}

// executeCommand runs the specified command with arguments and returns an error if it fails
func executeCommand(cmdName string, args []string) error {
	cmd := exec.Command(cmdName, args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command not found")
	}
	return nil
}