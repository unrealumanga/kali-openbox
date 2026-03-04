# NL2Cmd - Natural Language to Command Converter

A zsh/bash wrapper that converts natural language to shell commands using the ai-core daemon.

## Overview

NL2Cmd intercepts natural language commands prefixed with `?` or `ai`, sends them to the ai-core daemon (localhost:8080), and returns suggested commands for user verification before execution.

## Features

- **Safe by default**: Never executes commands automatically - always asks for confirmation
- **Two interfaces**: Zsh function integration + standalone Go binary
- **Flexible**: Works with any command the AI can suggest

## Installation

### 1. Build the Go binary

```bash
cd /path/to/nl2cmd
go build -o nl2cmd main.go
sudo mv nl2cmd /usr/local/bin/
```

### 2. Configure Zsh

Add to your `~/.zshrc`:

```zsh
# Load NL2Cmd functions
source /path/to/kali-openbox/tools/nl2cmd/nl2cmd.zsh

# Alias ? to ai for convenience
alias ?='ai '
```

Then reload your shell:
```bash
source ~/.zshrc
```

### 3. Ensure ai-core daemon is running

The ai-core daemon must be running on `localhost:8080`. Start it if needed:
```bash
ai-core &
```

## Usage

### Using the zsh function (recommended)

```bash
# Using 'ai' prefix
ai find all open ports in network 192.168.1.0/24

# Using '?' prefix (after adding alias)
? list all docker containers
? find files larger than 100MB in current directory
? show me the last 100 lines of /var/log/syslog
```

### Using the standalone binary

```bash
# Basic usage
nl2cmd "find all open ports in network 192.168.1.0/24"

# Execute immediately after confirmation
nl2cmd -e "list all running processes"

# Execute without confirmation
nl2cmd --execute "disk usage summary"
```

## Workflow

1. User enters natural language command (prefixed with `?` or `ai`)
2. NL2Cmd sends the prompt to ai-core daemon at localhost:8080
3. AI returns a suggested shell command
4. Command is displayed for user verification
5. User confirms (y/N) before execution
6. Command runs in the current shell

## Example Session

```
$ ? find all open ports in network 192.168.1.0/24
🤖 NL2Cmd: Converting to command...
   Prompt: find all open ports in network 192.168.1.0/24

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 Suggested command:

   nmap -sS 192.168.1.0/24

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Execute this command? [y/N] y
▶ Executing...
Starting Nmap...
```

## API Format

NL2Cmd sends POST requests to `http://localhost:8080/api/command`:

```json
{
  "prompt": "find all open ports in network 192.168.1.0/24",
  "context": "linux"
}
```

Expected response:
```json
{
  "command": "nmap -sS 192.168.1.0/24"
}
```

## Troubleshooting

### "Could not connect to ai-core daemon"
- Make sure ai-core is running: `ps aux | grep ai-core`
- Start it: `ai-core &`
- Check the port: `curl http://localhost:8080/health`

### Command not found
- Ensure `/usr/local/bin` is in your `$PATH`
- Or use full path: `/path/to/nl2cmd`

### Permission denied
- Make the binary executable: `chmod +x /usr/local/bin/nl2cmd`
