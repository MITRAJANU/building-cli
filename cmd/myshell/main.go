package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

const (
	cExit = "exit"
	cPwd  = "pwd"
	cCd   = "cd"
	cEcho = "echo"
	cType = "type"
	cCat  = "cat"
)

var (
	cBuiltins = []string{cExit, cPwd, cCd, cEcho, cType}
)

// absoluteBinpath returns the absolute path of a binary in the system PATH.
func absoluteBinpath(name string) string {
	paths := strings.Split(os.Getenv("PATH"), ":")

	for _, path := range paths {
		fullpath := filepath.Join(path, name)
		if _, err := os.Stat(fullpath); err == nil {
			return fullpath
		}
	}

	return ""
}

// buildString constructs a string from byte slice with specific trimming logic.
func buildString(s []byte, startIndex, lastIndex, spaceCheckIndex int, indicesToTrim []int) (original, sanitized string) {
	var output strings.Builder
	current := startIndex

	for _, index := range indicesToTrim {
		output.Write(s[current:index])
		current = index + 1
	}

	if current <= lastIndex {
		output.Write(s[current:lastIndex])
	}

	original = output.String()
	sanitized = original

	if spaceCheckIndex < len(s) && s[spaceCheckIndex] == ' ' {
		sanitized += " "
	}

	return original, sanitized
}

// parseCommandArg parses command arguments from a byte slice.
func parseCommandArg(arg []byte) (args []string, stringifiedArgs string) {
	var stringArgs strings.Builder
	var startIndex int
	var currentdelim byte
	var indicesToTrim []int

	maxIndex := len(arg) - 1

	for i := range arg {
		nextIndex := i + 1
		prevIndex := i - 1
		prevPrevIndex := prevIndex - 1

		switch arg[i] {
		case '\'':
			switch currentdelim {
			case '\'':
				original, sanitized := buildString(arg, startIndex, i, i+1, nil)
				stringArgs.WriteString(sanitized)
				args = append(args, original)
				startIndex = i + 1
				currentdelim = 0
			case 0:
				startIndex = i + 1
				currentdelim = '\''
			}
			continue

		case '"':
			switch currentdelim {
			case '"':
				if arg[prevIndex] != '\\' || (prevPrevIndex >= 0 && arg[prevPrevIndex] == '\\') {
					original, sanitized := buildString(arg, startIndex, i, i+1, indicesToTrim)
					stringArgs.WriteString(sanitized)
					args = append(args, original)
					startIndex = i + 1
					currentdelim = 0
					indicesToTrim = nil
					continue
				}
			case 0:
				if prevIndex < 0 || arg[prevIndex] != '\\' || (prevPrevIndex >= 0 && arg[prevPrevIndex] == '\\') {
					startIndex = i + 1
					currentdelim = '"'
					continue
				}
			}

		case ' ':
			if currentdelim == 0 {
				startIndex = i + 1
			}
			continue

		case '\\':
			if currentdelim == '\'' {
				continue
			}
			noDelimSkip := nextIndex <= maxIndex && currentdelim == 0 && (arg[nextIndex] == '\\' || arg[nextIndex] == '"' || arg[nextIndex] == ' ' || arg[nextIndex] == 'n')
			quaotationDelimSkip := nextIndex <= maxIndex && currentdelim == '"' && (arg[nextIndex] == '\\' || arg[nextIndex] == '"')

			if (noDelimSkip || quaotationDelimSkip) && (prevIndex < 0 || arg[prevIndex] != '\\') {
				indicesToTrim = append(indicesToTrim, i)
			}
		}

		if currentdelim == 0 && (nextIndex > maxIndex || arg[nextIndex] == ' ') {
			original, sanitized := buildString(arg, startIndex, nextIndex, nextIndex, indicesToTrim)
			stringArgs.WriteString(sanitized)
			args = append(args, original)
			startIndex = i + 1
			indicesToTrim = nil
		}
	}

	return args, stringArgs.String()
}

// executeCommand executes a command based on its type and arguments.
func executeCommand(commandType, commandArg string) string {
	switch commandType {
	case cExit:
		os.Exit(0)

	case cPwd:
		dir, err := os.Getwd()
		if err != nil {
			panic(fmt.Errorf("unable to obtain current working directory: %w", err))
		}
		return dir

	case cCd:
		if commandArg == "~" {
			commandArg = os.Getenv("HOME")
		}
		if err := os.Chdir(commandArg); err != nil {
			return fmt.Sprintf("cd: %v: No such file or directory", commandArg)
		}
		return ""

	case cEcho:
		_, output := parseCommandArg([]byte(commandArg))
		return output

	case cType:
		if slices.Contains(cBuiltins, commandArg) {
			return fmt.Sprintf("%v is a shell builtin", commandArg)
		}

		fullpath := absoluteBinpath(commandArg)
		if fullpath != "" {
			return fmt.Sprintf("%v is %v", commandArg, fullpath)
		}
		return fmt.Sprintf("%v: not found", commandArg)

	default:
		_, command := parseCommandArg([]byte(commandType))
		if fullpath := absoluteBinpath(command); fullpath != "" {
			args := []string{commandArg}
			if commandType == cCat {
				args, _ = parseCommandArg([]byte(commandArg))
			}
			cmd := exec.Command(fullpath, args...)
			output, err := cmd.Output()
			if err != nil {
				panic(fmt.Errorf("unable to run external program: %w", err))
			}
			return strings.TrimRightFunc(string(output), unicode.IsSpace)
		}
	}

	return fmt.Sprintf("%v: command not found", commandType)
}

// main function to run the shell.
func main() {
	for {
		fmt.Printf("$ ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(fmt.Errorf("unable to read input: %w", err))
		}

		command := input[:len(input)-1]
		args := make([]string, 2)

		var delim byte = ' '
		var stopIndex, continueIndex int
		var delimpresent bool

		for i := range command {
			if i == 0 {
				if command[i] == '\'' || command[i] == '"' {
					delim = command[i]
				}
				continue
			}

			if command[i] == delim {
				switch delim {
				case ' ':
					stopIndex = i
					continueIndex = i + 1

				default:
					stopIndex = i + 1
					continueIndex = stopIndex + 1
				}

				delimpresent = true
				break
			}
		}

		if !delimpresent {
			stopIndex = len(command)
			continueIndex = stopIndex
		}

		args[0] = command[:stopIndex]
		args[1] = command[continueIndex:]

		output := executeCommand(args[0], args[1])
		if output != "" {
			fmt.Println(output)
		}
	}
}
