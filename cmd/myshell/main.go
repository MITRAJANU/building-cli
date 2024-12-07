package main

import (

	"bufio"

	"fmt"

	"log"

	"os"

	"os/exec"

	"path"

	"regexp"

	"strconv"

	"strings"

)

func main() {

	for {

		// Uncomment this block to pass the first stage

		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input

		s, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {

			log.Fatal(err)

		}

		s = strings.Trim(s, "\r\n")

		var args []string

		command, argstr, _ := strings.Cut(s, " ")

		if strings.Contains(s, "\"") {

			re := regexp.MustCompile("\"(.*?)\"")

			args = re.FindAllString(s, -1)

			for i := range args {

				args[i] = strings.Trim(args[i], "\"")

			}

		} else if strings.Contains(s, "'") {

			re := regexp.MustCompile("'(.*?)'")

			args = re.FindAllString(s, -1)

			for i := range args {

				args[i] = strings.Trim(args[i], "'")

			}

		} else {

			args = strings.Fields(argstr)

			if strings.Contains(argstr, "\\") {

				re := regexp.MustCompile(`[^\\] +`)

				args = re.Split(argstr, -1)

				for i := range args {

					args[i] = strings.ReplaceAll(args[i], "\\", "")

				}

			} else {

				args = strings.Fields(argstr)

			}

		}

		switch command {

		case "cd":

			if args[0] == "~" {

				args[0] = os.Getenv("HOME")

			}

			if err := os.Chdir(args[0]); os.IsNotExist(err) {

				fmt.Println(command + ": " + args[0] + ": No such file or directory")

				break

			} else if err != nil {

				log.Fatal(err)

			}

		case "echo":

			fmt.Println(strings.Join(args, " "))

		case "exit":

			n, err := strconv.Atoi(args[0])

			if err != nil {

				log.Fatal(err)

			}

			os.Exit(n)

		case "pwd":

			dir, err := os.Getwd()

			if err != nil {

				log.Fatal(err)

			}

			fmt.Println(dir)

		case "type":

			var isBuiltin bool

			var isExecutable bool

			var cmdPath string

			builtin := []string{"cd", "echo", "exit", "pwd", "type"}

			for _, cmd := range builtin {

				if args[0] == cmd {

					isBuiltin = true

					break

				}

			}

			var separator string

			pathVar := os.Getenv("PATH")

			if strings.Contains(pathVar, ";") {

				separator = ";"

			} else {

				separator = ":"

			}

			dirs := strings.Split(pathVar, separator)

			for _, dir := range dirs {

				_, err = os.Stat(path.Join(dir, args[0]))

				if err == nil {

					isExecutable = true

					cmdPath = path.Join(dir, args[0])

					break

				}

			}

			if isBuiltin {

				fmt.Println(args[0] + " is a shell builtin")

			} else if isExecutable {

				fmt.Println(args[0] + " is " + cmdPath)

			} else {

				fmt.Println(args[0] + ": not found")

			}

		default:

			_, err = exec.LookPath(command)

			if err != nil {

				fmt.Println(command + ": command not found")

				break

			}

			cmd := exec.Command(command, args...)

			cmd.Stdin = os.Stdin

			cmd.Stdout = os.Stdout

			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {

				log.Fatal(err)

			}

		}

	}

}