package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var _ = fmt.Fprint

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

		input = strings.Trim(input, "\n")

		if input == "exit 0" {
			os.Exit(0)
		}

		args := parseInput(input)

		if len(args) == 0 {
			continue
		}

		command := args[0]
        args = args[1:]

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
                os.Exit(0)
            }
            os.Stdout.Write([]byte(absWdPath + "\n"))
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