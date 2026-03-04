# pod-sec: Ephemeral Security Tools Engine

An automatic container-based tool runner for OffTopic Linux that pulls and runs security tools on-demand as ephemeral containers.

## Features

- **On-demand tool execution**: Tools are pulled as containers only when needed
- **Automatic cleanup**: Containers are automatically removed after execution
- **Command not found interception**: Automatically catches unknown commands and runs them as containers
- **Current directory mounting**: Your working directory is mounted in the container
- **User context preservation**: Runs with your current user's UID/GID
- **Podman preferred, Docker fallback**: Automatically detects available container runtime

## Installation

1. Clone or copy this repository:
   ```bash
   cp -r /path/to/kali-openbox/tools/pod-sec ~/tools/
   ```

2. Build the Go binary (optional - it builds on first run):
   ```bash
   cd ~/tools/pod-sec
   go build -o pod-sec main.go
   ```

3. Source the wrapper in your shell rc file:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   source ~/tools/pod-sec/pod-sec.sh
   ```

4. Set the PODSEC_DIR environment variable if using a custom path:
   ```bash
   export PODSEC_DIR="$HOME/tools/pod-sec"
   ```

## Usage

### Direct invocation
```bash
pod-sec nmap -sV 192.168.1.1
pod-sec sqlmap -u http://target.com
pod-sec nikto -h http://target.com
```

### Command not found mode
After sourcing `pod-sec.sh`, simply run any mapped tool directly:
```bash
nmap -sV 192.168.1.1
hydra -L users.txt -P passwords.txt 192.168.1.1 ssh
```

## Available Tools

The tool mapping includes 90+ security tools including:

- **Network Scanning**: nmap, masscan
- **Web Vulnerability**: sqlmap, nikto, dirb, gobuster, wpscan, ffuf, nuclei
- **Password Attacks**: hydra, john, hashcat
- **Enumeration**: enum4linux, ldapsearch, smbclient
- **AD Attacks**: impacket, responder, evil-winrm, bloodhound, certipy
- **Privilege Escalation**: linpeas, winpeas, linux-exploit-suggester
- **Post-Exploitation**: mimikatz, rubeus, crackmapexec
- **API Testing**: postman, insomnia

See `tool-mapping.json` for the complete list.

## Configuration

### Custom tool mapping

Edit `tool-mapping.json` to add or modify tools:

```json
{
  "mytool": {
    "image": "myregistry/mytool:latest",
    "command": "mytool",
    "network": "host",
    "privileged": false,
    "volumes": ["/extra/path:/mnt"]
  }
}
```

### Tool mapping fields

| Field | Type | Description |
|-------|------|-------------|
| `image` | string | Container image to pull |
| `command` | string | Override the command to run (default: tool name) |
| `network` | string | Network mode (e.g., "host", "bridge") |
| `privileged` | bool | Run with privileged mode |
| `volumes` | array | Additional volumes to mount |

## How It Works

1. **Tool lookup**: When you request a tool, it checks `tool-mapping.json`
2. **Image pull**: Pulls the Alpine-based container image (if not already present)
3. **Container run**: Runs an interactive container with:
   - Current working directory mounted
   - User's UID/GID for proper file permissions
   - Home directory accessible
   - Passwd/group files for user resolution
4. **Cleanup**: Container is automatically removed (`--rm`) after exit

## Requirements

- Go 1.18+ (for building)
- Podman or Docker
- Linux environment (tested on OffTopic Linux)

## Troubleshooting

### Container runtime not found
Ensure podman or docker is installed and in your PATH:
```bash
which podman docker
```

### Permission denied
Make sure the pod-sec binary is executable:
```bash
chmod +x pod-sec
```

### Tool not found
Verify the tool is in `tool-mapping.json`:
```bash
grep "toolname" tool-mapping.json
```

## License

MIT License
