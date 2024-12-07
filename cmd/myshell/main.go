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
	"pwd":  "is a shell builtin",
	"cd":   "is a shell builtin",
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

		// Split the command into name and arguments, handling quotes
		cmdName, args := parseCommand(command)

		if cmdName == "exit" {
			exitCode := 0 // Default exit code
			if len(args) > 0 {
				fmt.Sscanf(args[0], "%d", &exitCode) // Get exit code if provided
			}
			os.Exit(exitCode) // Exit with the specified code
		}

        if cmdName == "echo" {
            handleEcho(args) // Handle echo command
            continue
        }

        if cmdName == "pwd" {
            handlePwd() // Handle pwd command
            continue
        }

        if cmdName == "cd" {
            handleCd(args) // Handle cd command
            continue
        }

        if cmdName == "type" {
            handleType(args) // Handle type command
            continue
        }

        // Execute external commands and handle errors.
        if err := executeExternalCommand(cmdName, args); err != nil {
            fmt.Printf("%s: %s\n", cmdName, err.Error())
        }
    }
}

// parseCommand splits the command string into the command name and arguments,
// handling both single and double quotes for literal values.
func parseCommand(input string) (string, []string) {
	var cmdName string
	var args []string

	var currentArg strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for _, char := range input {
	    switch char {
	    case '\'':
	        inSingleQuote = !inSingleQuote // Toggle single quote state.
	    case '"':
	        inDoubleQuote = !inDoubleQuote // Toggle double quote state.
	    case ' ':
	        if inSingleQuote || inDoubleQuote { // If inside any quotes, keep adding spaces to current argument.
	            currentArg.WriteRune(char)
	        } else { // If outside quotes, finalize current argument.
	            if currentArg.Len() > 0 {
	                args = append(args, currentArg.String())
	                currentArg.Reset()
	            }
	        }
	    case '\\':
	        if inDoubleQuote { // If inside double quotes, treat backslash as escape for next character.
	            currentArg.WriteRune(char) // Add backslash to current argument.
	        } else if inSingleQuote { 
	            currentArg.WriteRune(char) // Add backslash literally when inside single quotes.
	        }
	    default:
	        currentArg.WriteRune(char) // Add character to current argument.
	    }
	}

	if currentArg.Len() > 0 { // Add last argument if exists.
	    args = append(args, currentArg.String())
	}

	if len(args) > 0 {
	    cmdName = args[0]
	    args = args[1:] // Remove command name from arguments.
	}

	return cmdName, args
}

// handleEcho prints the provided arguments as a single string.
func handleEcho(args []string) {
	fmt.Println(strings.Join(args, " ")) // Join arguments with space and print.
}

// handlePwd prints the current working directory.
func handlePwd() {
	dir, err := os.Getwd() // Get current working directory.
	if err != nil {
        fmt.Fprintln(os.Stderr, "Error getting current directory:", err)
        return
    }
	fmt.Println(dir) // Print the current working directory.
}

// handleCd changes the current working directory.
func handleCd(args []string) {
	if len(args) != 1 {
        fmt.Println("cd: too many arguments")
        return
    }

    path := args[0]

    // Check for home directory shortcut (~)
    if path == "~" {
        homeDir := os.Getenv("HOME") // Get user's home directory from environment variable.
        if homeDir == "" {
            fmt.Println("cd: HOME not set")
            return
        }
        path = homeDir // Use home directory path.
    }

    // Change to the specified directory.
    if err := os.Chdir(path); err != nil {
        fmt.Printf("cd: %s: No such file or directory\n", path)
    }
}

// handleType determines how a command would be interpreted and prints the result.
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
		
        // Check if it's an executable in PATH.
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
        fullPath := filepath.Join(dir, cmd)
        if _, err := os.Stat(fullPath); err == nil {
            return fullPath, true // Found executable.
        }
    }
    return "", false // Not found.
}

// executeExternalCommand runs an external command with arguments and captures its output.
func executeExternalCommand(cmdName string, args []string) error {
    cmdPath, found := findExecutable(cmdName)
    if !found {
        return fmt.Errorf("command not found")
    }

    cmd := exec.Command(cmdPath, args...) // Create command with path and arguments.

    // Capture output from the command.
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("error executing command: %v", err)
    }

    fmt.Print(string(output)) // Print the output of the command.
    return nil
}