# Smart Recon Pipeline (auto-recon)

Automated reconnaissance orchestration tool that runs nmap, dirb, and DNS enumeration in sequence, then uses AI to summarize findings.

## Features

- **Sequential Tool Execution**: Runs nmap (quick scan) → dirb → DNS enumeration
- **Multiple Output Formats**: JSON/XML output from each tool
- **AI-Powered Summary**: Pipes results to ai-core for intelligent analysis
- **Color-coded Output**: Easy-to-read summary with ANSI colors

## Requirements

- Go 1.21+
- nmap
- dirb
- dnsenum
- ai-core (local AI service)

## Installation

```bash
cd tools/auto-recon
go build -o auto-recon main.go config.yaml
chmod +x recon.sh
```

## Configuration

Edit `config.yaml` to customize:

- Tool paths
- Nmap scan options
- Dirb wordlist
- AI model settings
- Output directory

## Usage

### Basic Usage

```bash
# Single target (IP or domain)
./recon.sh 192.168.1.1
./recon.sh example.com
```

### Shell Wrapper

```bash
./recon.sh <target>
```

### Direct Binary

```bash
./auto-recon <target>
```

### Multiple Targets

```bash
./recon.sh 192.168.1.1,example.com
```

## Output

The tool creates a timestamped directory in `/tmp/recon/` containing:

- `nmap_results.xml` - Nmap scan results
- `dirb_results.txt` - Directory enumeration results  
- `dns_results.xml` - DNS enumeration results
- `recon_full_report.json` - Complete JSON report

## AI Summary

The AI analyzes and reports:

1. **Open Ports & Services** - Identified services with versions
2. **Potential Vulnerabilities** - Service-specific security concerns
3. **Interesting Directories** - Found paths from dirb
4. **DNS Findings** - DNS records and subdomains
5. **Recommendations** - Next steps for security assessment

## Example Output

```
[+] Smart Recon Pipeline Started for: example.com
[*] Output directory: /tmp/recon/example.com_20240315_143022

[*] Phase 1: Nmap Quick Scan
[*] Phase 2: Directory Enumeration (dirb)
[*] Phase 3: DNS Enumeration
[*] Phase 4: AI Analysis & Summary

=== AI Security Analysis ===
[Summary from AI...]

[+] Reconnaissance complete!
[*] Full report: /tmp/recon/example.com_20240315_143022/recon_full_report.json
```

## Color Legend

- Green: Open ports, successful discoveries
- Yellow: Warnings, versions
- Red: Errors
- Cyan: Phase indicators
- Blue: Section headers

## License

MIT License
