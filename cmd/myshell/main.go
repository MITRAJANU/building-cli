package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ErrNotFound    = errors.New("Error not found")
	ErrInvalidPath = errors.New("Invalid Path")
)

type Repl struct {
	commandMap  map[string]CommandHandler
	currentPath string
	running     bool
}

type Command struct {
	handler CommandHandler
	command string
}

func initRepl(cmdMap map[string]CommandHandler) *Repl {
	dir, err := getWorkingDirectory()
	if err != nil {
		log.Fatal(err)
	}
	return &Repl{
		running:     true,
		commandMap:  cmdMap,
		currentPath: dir,
	}
}

func (r *Repl) changeCurrentPath(newPath string) {
	r.currentPath = newPath
}

func (repl *Repl) stop() {
	repl.running = false
}

func (repl *Repl) setCommands(cmdMap map[string]CommandHandler) {
	repl.commandMap = cmdMap
}

func initCommands() []Command {
	return []Command{
		{handler: handleEcho, command: "echo"},
		{handler: handleExit, command: "exit"},
		{handler: handleType, command: "type"},
		{handler: handlePwd, command: "pwd"},
		{handler: handleCd, command: "cd"},
	}
}

type CommandHandler func(r *Repl, arguments []string)

func getWorkingDirectory() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func parseCommand(commandStr string) (string, []string) {
	var cmdName string
	var args []string

	inSingleQuote := false
	inDoubleQuote := false

	var currentArg strings.Builder

	for i, character := range commandStr {
		char := string(character)

		if char == "'" && !inDoubleQuote {
			inSingleQuote = !inSingleQuote // Toggle single quote state
			if !inSingleQuote { // If closing single quote, add argument to args
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
			continue
		} else if char == "\"" && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote // Toggle double quote state
			if !inDoubleQuote { // If closing double quote, add argument to args
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
			continue
		} else if char == "\\" && inSingleQuote { // Preserve backslash in single quotes
			currentArg.WriteString(char)
			continue
		}

		if char == " " && !inSingleQuote && !inDoubleQuote { // Split on spaces if not in quotes
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
			continue
		}

		currentArg.WriteString(char) // Add character to current argument
	}

	if currentArg.Len() > 0 { // Add last argument if exists after loop ends
		args = append(args, currentArg.String())
	}

	if len(args) > 0 {
		cmdName = args[0] // The first element is the command name
		args = args[1:]   // The rest are arguments
	}

	return cmdName, args
}

func findExecutableFile(directory string, fileName string) (string, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories for this simple example
		}
		
        filePath := filepath.Join(directory, file.Name())
        if file.Name() != fileName {
            continue 
        }

        info, err := os.Stat(filePath)
        if err != nil || info.Mode().Perm()&0o111 == 0 { 
            continue 
        }

        return filePath, nil 
    }
    return "", ErrNotFound 
}

func main() {
    commands := initCommands()
    cmdMap := make(map[string]CommandHandler)
    for _, cmd := range commands {
        cmdMap[cmd.command] = cmd.handler 
    }
    
    repl := initRepl(cmdMap)

    for repl.running {
        fmt.Fprint(os.Stdout, "$ ")

        commandStr, err := bufio.NewReader(os.Stdin).ReadString('\n')
        if err != nil {
            log.Fatalf("Error: %e", err)
        }

        commandStr = strings.TrimSpace(commandStr)

        command, args := parseCommand(commandStr)

        handler, ok := cmdMap[command]
        if !ok {
            fmt.Fprintf(os.Stdout, "%s: not found\n", command)
            continue 
        }

        handler(repl, args)
    }
}

// Handle echo command with preserved arguments.
func handleEcho(_ *Repl, arguments []string) {
	echoStr := ""
	if len(arguments) >= 1 {
	    echoStr = strings.Join(arguments, " ") + "\n"
    }
	fmt.Fprintf(os.Stdout, "%s", echoStr)
}

// Other handlers (handleExit, handleType etc.) remain unchanged.