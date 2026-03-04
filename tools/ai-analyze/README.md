# ai-analyze

AI-powered security artifact analyzer for OffTopic Linux.

Analyzes log files, pcap files, and other security artifacts using AI to identify credentials, CVE signatures, anomalies, and indicators of compromise (IOCs).

## Features

- **File Type Detection**: Automatically detects file types (.log, .pcap, .json, .txt, .xml)
- **Large File Handling**: Chunks large files for processing
- **Security Analysis**: Identifies:
  - Credentials (API keys, passwords, tokens, secrets)
  - CVE Signatures (vulnerability patterns, CVE references)
  - Anomalies (unusual patterns, failed auth attempts, suspicious activity)
  - IOCs (IP addresses, domains, hashes, URLs)
- **AI-Powered**: Uses ai-core for intelligent analysis
- **Multiple Output Formats**: Console output and JSON report files

## Requirements

- Go 1.21+
- ai-core (CLI tool or HTTP endpoint)
- yaml.v3 Go package

## Installation

```bash
cd tools/ai-analyze
go build -o ai-analyze main.go
```

Or use the shell wrapper:

```bash
chmod +x ai-analyze.sh
./ai-analyze.sh <file>
```

## Configuration

Edit `config.yaml` to customize:

```yaml
ai_core:
  endpoint: "http://localhost:8080/analyze"  # or use CLI mode
  model: "default"
  temperature: 0.3
  max_tokens: 4000
  timeout: 120

analysis:
  chunk_size: 50000
  max_file_size: 104857600  # 100MB

output:
  report_dir: "./reports"
  json_format: false
```

## Usage

```bash
# Basic usage
./ai-analyze <file>

# With output file
./ai-analyze -o report.json access.log

# Disable colors
./ai-analyze --no-color server.log

# Show help
./ai-analyze --help
```

## Supported File Types

| Extension | Type    |
|-----------|---------|
| .log      | log     |
| .pcap     | pcap    |
| .pcapng   | pcap    |
| .json     | json    |
| .txt      | text    |
| .xml      | xml     |

## Shell Completion

### Bash

```bash
source completion/bash_completion.sh
```

Or copy to bash completion directory:
```bash
sudo cp completion/bash_completion.sh /etc/bash_completion.d/ai-analyze
```

### Zsh

Add to your `.zshrc`:
```zsh
fpath=(path/to/ai-analyze/completion $fpath)
autoload -Uz _ai_analyze
```

Or copy to zsh completion directory:
```bash
sudo cp completion/zsh_completion.sh /usr/share/zsh/site-functions/_ai-analyze
```

## Output Example

```
[2024-01-15 10:30:45] AI Security Analysis
[*] File: access.log
[*] Type: log | Size: 2.5 MB | Chunks: 1

[CRITICAL] SEVERITY: High

[!] Credentials Found (2)
  [!] Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  [!] Basic dXNlcjpwYXNzMTIz

[!] CVE Signatures (1)
  [*] CVE-2023-1234 - SQL injection attempt in /admin

[*] Anomalies (3)
  - Multiple failed authentication attempts from 192.168.1.100
  - Unusual traffic spike detected
  - Privilege escalation attempt

[*] Indicators of Compromise (2)
  [*] [ip] 10.0.0.50
  [*] [domain] malicious-site.com

[*] Summary:
The log file shows evidence of a SQL injection attack targeting the /admin endpoint...
```

## Integration with ai-core

The tool supports two modes:

1. **HTTP Mode**: If `ai_core.endpoint` starts with `http`, it sends POST requests
2. **CLI Mode**: Falls back to calling `ai-core` command-line tool with arguments

Both modes use the same system prompt for consistent analysis.
