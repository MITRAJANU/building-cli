package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

func parseInput(input string) []string {
	var args []string
	var currentArg strings.Builder

	inSingleQuotes := false
	inDoubleQuotes := false
	escapeCh := false

	for _, char := range input {
		if escapeCh {
			currentArg.WriteRune(char)
			escapeCh = false
			continue
		}

		switch char {
		case '\\':
			if inDoubleQuotes {
				escapeCh = true // Escape next character in double quotes
			} else {
				currentArg.WriteRune(char) // Literal backslash in single quotes
			}
		case '\'':
			if !inDoubleQuotes {
				inSingleQuotes = !inSingleQuotes // Toggle single-quote mode
			} else {
				currentArg.WriteRune(char)
			}
		case '"':
			if !inSingleQuotes {
				inDoubleQuotes = !inDoubleQuotes // Toggle double-quote mode
			} else {
				currentArg.WriteRune(char)
			}
		case ' ':
			if inSingleQuotes || inDoubleQuotes {
				currentArg.WriteRune(char) // Space inside quotes is part of the argument
			} else if currentArg.Len() > 0 {
				args = append(args, currentArg.String()) // End of argument
				currentArg.Reset()
			}
		default:
			currentArg.WriteRune(char) // Add character to the current argument
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String()) // Add final argument
	}

	return args
}

func main() {
	builtInCmds := []string{"echo", "exit", "type", "pwd", "cd"}

	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatalln(err.Error())
		}

        input = strings.TrimSpace(input)

        if input == "exit 0" {
            os.Exit(0)
        }

        parts := parseInput(input)

        if len(parts) == 0 {
            continue
        }

        command := parts[0]
        args := parts[1:]

        switch command {
        case "cd":
            if len(args) == 0 {
                fmt.Println("cd: missing argument")
                continue
            }
            dirPath := args[0]
            if dirPath == "~" {
                dirPath = os.Getenv("HOME")
            }
            err := os.Chdir(dirPath)
            if err != nil {
                fmt.Printf("cd: %s: No such file or directory\n", dirPath)
            }
        case "pwd":
            absWdPath, err := os.Getwd()
            if err != nil {
                log.Fatal(err)
            }
            fmt.Println(absWdPath)
        case "echo":
            fmt.Println(strings.Join(args, " "))
        case "type":
            if len(args) < 1 {
                fmt.Println("type: missing argument")
                continue
            }
            cmd := args[0]
            if slices.Contains(builtInCmds, cmd) {
                fmt.Printf("%s is a shell builtin\n", cmd)
            } else {
                paths := strings.Split(os.Getenv("PATH"), ":")
                foundedPath := ""
                for _, path := range paths {
                    fp := filepath.Join(path, cmd)
                    fileInfo, err := os.Stat(fp)
                    if err == nil && !fileInfo.IsDir() { 
                        foundedPath = fp
                        break
                    }
                }
                if foundedPath != "" {
                    fmt.Printf("%s is %s\n", cmd, foundedPath)
                } else {
                    fmt.Printf("%s: not found\n", cmd)
                }
            }
        default:
            execCmd := exec.Command(command, args...)
            execCmd.Stderr = os.Stderr
            execCmd.Stdin = os.Stdin
            execCmd.Stdout = os.Stdout

            err := execCmd.Run()
            if err != nil {
                fmt.Printf("%s: command not found\n", command)
            }
        }
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

// Handle exit command.
func handleExit(r *Repl, arguments []string) {
	if len(arguments) < 1 { return }
	if arguments[0] == "0" { r.stop() }
}

// Handle type command.
func handleType(r *Repl, arguments []string) { /* Implementation here */ }

// Handle pwd command.
func handlePwd(r *Repl, _ []string) { /* Implementation here */ }

// Handle cd command.
func handleCd(r *Repl, args []string) { /* Implementation here */ }

// Other helper functions can be added as needed.