package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Request struct {
	Prompt  string `json:"prompt"`
	Context string `json:"context"`
}

type Response struct {
	Command string `json:"command"`
	Error   string `json:"error,omitempty"`
}

var (
	execute   bool
	apiURL    string
	shellPath string
)

func main() {
	apiURL = "http://localhost:8080/api/command"
	shellPath = "/bin/zsh"
	if _, err := os.Stat(shellPath); err != nil {
		shellPath = "/bin/bash"
	}

	args := os.Args[1:]
	for i, arg := range args {
		if arg == "-e" || arg == "--execute" {
			execute = true
			args = append(args[:i], args[i+1:]...)
			break
		}
		if arg == "-h" || arg == "--help" {
			printUsage()
			os.Exit(0)
		}
		if (arg == "-u" || arg == "--url") && i+1 < len(args) {
			apiURL = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		}
		if (arg == "-s" || arg == "--shell") && i+1 < len(args) {
			shellPath = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	prompt := strings.Join(args, " ")

	req := Request{
		Prompt:  prompt,
		Context: "linux",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create request: %v\n", err)
		os.Exit(1)
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create HTTP request: %v\n", err)
		os.Exit(1)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not connect to ai-core daemon at %s\n", apiURL)
		fmt.Fprintf(os.Stderr, "   Make sure the daemon is running.\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid JSON response: %v\n", err)
		os.Exit(1)
	}

	if apiResp.Error != "" {
		fmt.Fprintf(os.Stderr, "Error from ai-core: %s\n", apiResp.Error)
		os.Exit(1)
	}

	if apiResp.Command == "" {
		fmt.Fprintf(os.Stderr, "Error: No command returned from ai-core\n")
		os.Exit(1)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Suggested command:")
	fmt.Println()
	fmt.Printf("   %s\n", apiResp.Command)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if execute {
		fmt.Print("Executing...\n\n")
		runCmd := exec.Command(shellPath, "-c", apiResp.Command)
		runCmd.Stdin = os.Stdin
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr
		err := runCmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
			os.Exit(1)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Execute this command? [y/N] ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			fmt.Print("Executing...\n\n")
			runCmd := exec.Command(shellPath, "-c", apiResp.Command)
			runCmd.Stdin = os.Stdin
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
			err := runCmd.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Command not executed.")
		}
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `NL2Cmd - Natural Language to Command Converter

Usage: nl2cmd [options] "<natural language command>"

Options:
  -e, --execute    Execute the command after user confirmation
  -u, --url URL    API URL (default: http://localhost:8080/api/command)
  -s, --shell SHELL Shell to use (default: /bin/zsh or /bin/bash)
  -h, --help       Show this help message

Examples:
  nl2cmd "find all open ports in network 192.168.1.0/24"
  nl2cmd -e "list all running docker containers"
  nl2cmd --execute "find files modified in last 24 hours"
`)
}
