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

		var command string
		var args []string

		command, args = parseCommand(s)

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
	command, argstr, _ := strings.Cut(input, " ")
	var singleQuote, doubleQuote, backslash bool
	var arg string
	var args []string

	for _, r := range argstr {
		switch r {
		case '\'':
			if backslash && doubleQuote {
				arg += "\\"
			}
			if backslash || doubleQuote {
				arg += string(r)
			} else {
				singleQuote = !singleQuote
			}
			backslash = false

		case '"':
			if backslash || singleQuote {
				arg += string(r)
			} else {
				doubleQuote = !doubleQuote
			}
			backslash = false

		case '\\':
			if backslash || singleQuote {
				arg += string(r)
				backslash = false
			} else {
				backslash = true
			}

		case ' ':
			if backslash && doubleQuote {
				arg += "\\"
			}
			if backslash || singleQuote || doubleQuote {
				arg += string(r)
			} else if arg != "" {
				args = append(args, arg)
				arg = ""
			}
			backslash = false

		default:
			if doubleQuote && backslash {
				arg += "\\"
			}
			arg += string(r)
			backslash = false
		}
	}

	if arg != "" {
		args = append(args, arg)
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
