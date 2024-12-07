package main

import (

	"bufio"

	"errors"

	"fmt"

	"log"

	"os"

	"os/exec"

	"path/filepath"

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

	commands := []Command{}

	commands = append(commands, Command{

		handler: handleEcho,

		command: "echo",

	})

	commands = append(commands, Command{

		handler: handleExit,

		command: "exit",

	})

	commands = append(commands, Command{

		handler: handleType,

		command: "type",

	})

	commands = append(commands, Command{

		handler: handlePwd,

		command: "pwd",

	})

	commands = append(commands, Command{

		handler: handleCd,

		command: "cd",

	})

	return commands

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

	quoteFlag := false

	doubleQuoteFlag := false

	commandSplit := []string{""}

	for i, characterRune := range commandStr {

		character := string(characterRune)

		slashFlag := false

		if i > 0 {

			slashFlag = string(commandStr[i-1]) == "\\"

		}

		if character == string('"') && !slashFlag {

			doubleQuoteFlag = !doubleQuoteFlag

		} else if character == " " && !quoteFlag && !doubleQuoteFlag && !slashFlag {

			commandSplit = append(commandSplit, "")

		} else if character == "'" && !doubleQuoteFlag && !slashFlag {

			quoteFlag = !quoteFlag

		} else if character == "\\" && !slashFlag && !doubleQuoteFlag {

		} else if character == "\\" && !slashFlag && !doubleQuoteFlag && !quoteFlag {

			continue

		} else {

			currentIndex := len(commandSplit) - 1

			commandSplit[currentIndex] += character

		}

	}

	command := commandSplit[0]

	commandArgs := []string{}

	if len(commandSplit) > 1 {

		commandArgs = commandSplit[1:]

	}

	commandArgs = removeEmpty(commandArgs)

	return command, commandArgs

}

func findExecuteableFile(directory string, fileName string) (string, error) {

	files, err := os.ReadDir(directory)

	if err != nil {

		return "", err

	}

	for _, file := range files {

		if file.IsDir() {

			// filePath, err := findExecuteableFile(directory+"/"+file.Name(), fileName)

			// if err != nil {

			// 	return filePath, nil

			// }

			continue

		}

		filePath := filepath.Join(directory, file.Name())

		if file.Name() != fileName {

			continue

		}

		// Get file info

		info, err := os.Stat(filePath)

		if err != nil {

			fmt.Printf("Error getting file info: %v\n", err)

			continue

		}

		// Check if the file is executable

		if info.Mode().Perm()&0o111 == 0 {

			continue

		}

		return filePath, nil

	}

	return "", ErrNotFound

}

func fileExists(directory string, fileName string) (string, error) {

	filePath := filepath.Join(directory, fileName)

	_, err := os.Stat(filePath)

	if err == nil {

		return filePath, nil

	}

	return "", ErrNotFound

}

func getCommandMap(commands []Command) map[string]CommandHandler {

	commandMap := map[string]CommandHandler{}

	for _, command := range commands {

		commandMap[command.command] = command.handler

	}

	return commandMap

}

func programExistsInPath(fileName string) (string, error) {

	paths := strings.Split(os.Getenv("PATH"), ":")

	for _, path := range paths {

		filePath := filepath.Join(path, fileName)

		_, err := os.Stat(filePath)

		if err == nil {

			return filePath, nil

		}

	}

	return "", ErrNotFound

}

func executeProgram(filePath string, args []string) error {

	cmd := exec.Command(filePath, args...)

	// Run the command and capture the output

	cmd.Stderr = os.Stderr

	cmd.Stdout = os.Stdout

	return cmd.Run()

}

func main() {

	// Uncomment this block to pass the first stage

	commands := initCommands()

	cmdMap := getCommandMap(commands)

	repl := initRepl(cmdMap)

	for repl.running {

		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input

		commandStr, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {

			log.Fatalf("Error: %e", err)

		}

		commandStr = commandStr[:len(commandStr)-1]

		command, args := parseCommand(commandStr)

		handler, ok := cmdMap[command]

		if !ok {

			// try find program otherwise not found

			filePath, err := programExistsInPath(command)

			if err == nil {

				err = executeProgram(filePath, args)

				if err == nil {

					continue

				}

			}

			fmt.Fprintf(os.Stdout, "%s: not found\n", command)

			continue

		}

		handler(repl, args)

	}

}

func handleEcho(_ *Repl, arguments []string) {

	echoStr := ""

	if len(arguments) >= 1 {

		echoStr = strings.Join(arguments, " ") + "\n"

	}

	fmt.Fprintf(os.Stdout, "%s", echoStr)

}

func handleType(r *Repl, arguments []string) {

	if len(arguments) < 1 {

		return

	}

	cmd := arguments[0]

	_, ok := r.commandMap[cmd]

	if ok {

		fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", cmd)

		return

	}

	directories := strings.Split(os.Getenv("PATH"), ":")

	filePath, err := "", ErrNotFound

	for _, directory := range directories {

		filePath, err = fileExists(directory, cmd)

		if err == nil {

			break

		}

	}

	if err != nil {

		fmt.Fprintf(os.Stdout, "%s: not found\n", cmd)

		return

	}

	fmt.Fprintf(os.Stdout, "%s is %s\n", cmd, filePath)

}

func handleExit(r *Repl, arguments []string) {

	if len(arguments) < 1 {

		return

	}

	if arguments[0] == "0" {

		r.stop()

	}

}

func handlePwd(r *Repl, _ []string) {

	fmt.Fprintf(os.Stdout, "%s\n", r.currentPath)

}

func handleCd(r *Repl, args []string) {

	if len(args) < 1 {

		return

	}

	newPath := args[0]

	paths := strings.Split(newPath, "/")

	newPaths := strings.Split(r.currentPath, "/")

	newPaths = removeEmpty(newPaths)

	var err error = nil

out:

	for i, path := range paths {

		switch path {

		case "":

			if i == 0 {

				newPaths = []string{}

			}

			continue

		case ".":

			continue

		case "~":

			newPaths = removeEmpty(strings.Split(os.Getenv("HOME"), "/"))

		case "..":

			if len(newPaths) < 1 {

				err = ErrInvalidPath

				break out

			}

			newPaths = newPaths[:len(newPaths)-1]

		default:

			newPaths = append(newPaths, path)

		}

	}

	if err != nil {

		fmt.Fprintf(os.Stdout, "cd: %s: No such file or directory\n", newPath)

		return

	}

	newPath = "/" + strings.Join(newPaths, "/")

	_, err = os.Stat(newPath)

	if err != nil {

		fmt.Fprintf(os.Stdout, "cd: %s: No such file or directory\n", newPath)

		return

	}

	r.changeCurrentPath(newPath)

}

func removeEmpty(oldList []string) []string {

	newList := []string{}

	for _, li := range oldList {

		if li != "" {

			newList = append(newList, li)

		}

	}

	return newList

}