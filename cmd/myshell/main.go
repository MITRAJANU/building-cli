package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

// List of built-in commands
var builtins = map[string]string{
	"echo": "is a shell builtin",
	"exit": "is a shell builtin",
	"type": "is a shell builtin",
}

func main() {
	for {
		// Prompt for user input
		fmt.Print("$ ")

		// Read user input
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
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

		if cmdName == "exit" {
			exitCode := 0 // Default exit code
			if len(args) > 1 {
				fmt.Sscanf(args[1], "%d", &exitCode) // Get exit code if provided
			}
			os.Exit(exitCode) // Exit with the specified code
		}

		if cmdName == "echo" {
			handleEcho(args[1:]) // Handle echo command
			continue
		}

		if cmdName == "type" {
			handleType(args[1:]) // Handle type command
			continue
		}

		// Execute other commands and handle errors
		if err := executeCommand(cmdName, args[1:]); err != nil {
			fmt.Printf("%s: %s\n", cmdName, err.Error())
		}
	}
}

// handleEcho prints the provided arguments as a single string
func handleEcho(args []string) {
	fmt.Println(strings.Join(args, " ")) // Join arguments with space and print
}

// handleType determines how a command would be interpreted and prints the result
func handleType(args []string) {
	if len(args) == 0 {
		fmt.Println("type: missing argument")
		return
	}

	for _, arg := range args {
		if desc, exists := builtins[arg]; exists {
			fmt.Printf("%s %s\n", arg, desc)
			continue
		}
		
        // Check if it's an executable in PATH
        path, found := findExecutable(arg)
        if found {
            fmt.Printf("%s is %s\n", arg, path)
        } else {
            fmt.Printf("%s: not found\n", arg)
        }
    }
}

// findExecutable searches for an executable in the directories listed in PATH.
func findExecutable(cmd string) (string, bool) {
	pathEnv := os.Getenv("PATH")
	directories := strings.Split(pathEnv, ":")

	for _, dir := range directories {
        // Construct full path for the executable
        fullPath := filepath.Join(dir, cmd)
        
        // Check if it exists and is executable
        if _, err := os.Stat(fullPath); err == nil {
            return fullPath, true // Found executable
        }
    }
    return "", false // Not found
}

// executeCommand runs the specified command with arguments and returns an error if it fails.
func executeCommand(cmdName string, args []string) error {
	cmd := exec.Command(cmdName, args...)
	err := cmd.Run()
	if err != nil {
	    return fmt.Errorf("command not found")
    }
    return nil
}