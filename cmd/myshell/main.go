package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func main() {
	for {
		// Prompt for user input
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		s, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		s = strings.Trim(s, "\r\n")

		command, args := parseCommand(s)

		switch command {
		case "cd":
			handleCD(args)
		case "echo":
			fmt.Println(strings.Join(args, " "))
		case "exit":
			handleExit(args)
		case "pwd":
			handlePWD()
		case "type":
			handleType(args)
		default:
			executeCommand(command, args)
		}
	}
}

func parseCommand(input string) (string, []string) {
	var command string
	var args []string

	if strings.HasPrefix(input, "'") || strings.HasPrefix(input, "\"") {
		input = strings.Trim(input, "'\"")
	}

	commandEnd := strings.IndexAny(input, " ")
	if commandEnd == -1 {
		command = input
	} else {
		command = input[:commandEnd]
		argstr := input[commandEnd+1:]

		var singleQuote, doubleQuote bool
		var arg string

		for i := 0; i < len(argstr); i++ {
			r := rune(argstr[i])
			switch r {
			case '\'':
				singleQuote = !singleQuote
			case '"':
				doubleQuote = !doubleQuote
			case ' ':
				if singleQuote || doubleQuote {
					arg += string(r)
				} else if arg != "" {
					args = append(args, arg)
					arg = ""
				}
			default:
				arg += string(r)
			}
		}

		if arg != "" {
			args = append(args, arg)
		}
	}

	return command, args
}

func handleCD(args []string) {
	if len(args) == 0 {
		fmt.Println("cd: missing argument")
		return
	}
	if args[0] == "~" {
		args[0] = os.Getenv("HOME")
	}
	if err := os.Chdir(args[0]); os.IsNotExist(err) {
		fmt.Println("cd: " + args[0] + ": No such file or directory")
	} else if err != nil {
		log.Fatal(err)
	}
}

func handleExit(args []string) {
	n := 0
	if len(args) > 0 {
		var err error
		n, err = strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(n)
}

func handlePWD() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
}

func handleType(args []string) {
	if len(args) == 0 {
		fmt.Println("type: missing argument")
		return
	}

	builtin := []string{"cd", "echo", "exit", "pwd", "type"}
	isBuiltin := false

	for _, cmd := range builtin {
		if args[0] == cmd {
			isBuiltin = true
			break
		}
	}

	pathVar := os.Getenv("PATH")
	var cmdPath string

	for _, dir := range strings.Split(pathVar, ":") { // Assuming ':' as path separator for Unix-like systems.
		if _, err := os.Stat(path.Join(dir, args[0])); err == nil {
			cmdPath = path.Join(dir, args[0])
			break
		}
	}

	if isBuiltin {
		fmt.Println(args[0] + " is a shell builtin")
	} else if cmdPath != "" {
		fmt.Println(args[0] + " is " + cmdPath)
	} else {
		fmt.Println(args[0] + ": not found")
	}
}

func executeCommand(command string, args []string) {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
