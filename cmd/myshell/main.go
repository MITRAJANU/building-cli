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

			if inSingleQuotes {

				currentArg.WriteRune(char) // Literal backslash in single quotes

			} else {

				escapeCh = true // Escape next character

				// Within double code, add backslash character

				if inDoubleQuotes {

					currentArg.WriteRune(char)

				}

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

		// Wait for user input

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {

			log.Fatalln(err.Error())

		}

		// Raw user input

		input = strings.Trim(input, "\n")

		if input == "exit 0" {

			os.Exit(0)

		}

		parts := parseInput(input)

		if len(parts) == 0 {

			continue

		}

		// Shell command

		command := parts[0]

		args := parts[1:]

		switch command {

		case "cd":

			// Get dir path

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

			// Add new line character

			fmt.Println(strings.Join(args, " "))

		case "type":

			args := parts[1]

			paths := strings.Split(os.Getenv("PATH"), ":")

			foundedPath := ""

			if isContain := slices.Contains(builtInCmds, args); isContain {

				fmt.Printf("%s is a shell builtin\n", args)

			} else {

				for _, path := range paths {

					fp := filepath.Join(path, args)

					fileInfo, err := os.Stat(fp)

					if err != nil {

						continue

					}

					if fileInfo.Name() == args {

						foundedPath = fp

						break

					}

				}

				if foundedPath != "" {

					fmt.Printf("%s is %s\n", args, foundedPath)

				} else {

					fmt.Printf("%s: not found\n", args)

				}

			}

		default:

			// removedSingleQuoteArr := removeSingleQuote(parts[1:], "'")

			execCmd := exec.Command(command, args...)

			// Define how to handle error, input and output for external executable packages

			execCmd.Stderr = os.Stderr

			execCmd.Stdin = os.Stdin

			execCmd.Stdout = os.Stdout

			// Execute external program here

			err := execCmd.Run()

			if err != nil {

				fmt.Printf("%s: command not found\n", command)

			}

		}

	}

}