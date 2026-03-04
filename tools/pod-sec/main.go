package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type ToolMapping map[string]ToolConfig

type ToolConfig struct {
	Image         string   `json:"image"`
	BaseImage     string   `json:"base_image,omitempty"`
	Command       string   `json:"command,omitempty"`
	Privileged    bool     `json:"privileged,omitempty"`
	Network       string   `json:"network,omitempty"`
	Volumes       []string `json:"volumes,omitempty"`
}

var (
	mappingFile = "tool-mapping.json"
	containerRuntime = "podman"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <tool-name> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Error: No tool specified\n")
		os.Exit(1)
	}

	toolName := os.Args[1]
	args := os.Args[2:]

	mapping, err := loadMapping(mappingFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tool mapping: %v\n", err)
		os.Exit(1)
	}

	toolConfig, exists := mapping[toolName]
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Tool '%s' not found in mapping\n", toolName)
		fmt.Fprintf(os.Stderr, "Available tools: %s\n", strings.Join(getToolNames(mapping), ", "))
		os.Exit(127)
	}

	if err := runTool(toolName, toolConfig, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error running tool: %v\n", err)
		os.Exit(1)
	}
}

func loadMapping(filename string) (ToolMapping, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var mapping ToolMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, err
	}

	return mapping, nil
}

func getToolNames(mapping ToolMapping) []string {
	names := make([]string, 0, len(mapping))
	for name := range mapping {
		names = append(names, name)
	}
	return names
}

func detectRuntime() string {
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman"
	}
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker"
	}
	return ""
}

func runTool(toolName string, config ToolConfig, args []string) error {
	runtime := detectRuntime()
	if runtime == "" {
		return fmt.Errorf("neither podman nor docker found in PATH")
	}
	containerRuntime = runtime

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	userInfo, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	pullArgs := []string{"pull", config.Image}
	fmt.Printf("Pulling container image: %s\n", config.Image)
	pullCmd := exec.Command(runtime, pullArgs...)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	containerName := fmt.Sprintf("pod-sec-%s-%d", toolName, time.Now().UnixNano())

	runArgs := []string{
		"run",
		"--rm",
		"--name", containerName,
		"-it",
		"--workdir", cwd,
		"-v", cwd + ":" + cwd,
		"-v", "/etc/passwd:/etc/passwd:ro",
		"-v", "/etc/group:/etc/group:ro",
		"-e", "HOME=" + userInfo.HomeDir,
		"-e", "USER=" + userInfo.Username,
		"-e", "UID=" + userInfo.Uid,
		"-e", "GID=" + userInfo.Gid,
	}

	if config.Network != "" {
		runArgs = append(runArgs, "--network", config.Network)
	}

	if config.Privileged {
		runArgs = append(runArgs, "--privileged")
	}

	for _, vol := range config.Volumes {
		runArgs = append(runArgs, "-v", vol)
	}

	runArgs = append(runArgs, config.Image)

	toolCmd := config.Command
	if toolCmd == "" {
		toolCmd = toolName
	}
	runArgs = append(runArgs, toolCmd)
	runArgs = append(runArgs, args...)

	fmt.Printf("Running: %s %s\n", runtime, strings.Join(runArgs[2:], " "))
	runCmd := exec.Command(runtime, runArgs...)
	runCmd.Stdin = os.Stdin
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Dir = cwd

	if err := runCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}
