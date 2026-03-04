package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	MaxChunkSize = 50000
)

type Config struct {
	AiCore struct {
		Endpoint   string  `yaml:"endpoint"`
		Model      string  `yaml:"model"`
		Temperature float64 `yaml:"temperature"`
		MaxTokens  int     `yaml:"max_tokens"`
		Timeout    int     `yaml:"timeout"`
	} `yaml:"ai_core"`
	Analysis struct {
		ChunkSize   int  `yaml:"chunk_size"`
		MaxFileSize int64 `yaml:"max_file_size"`
	} `yaml:"analysis"`
	Output struct {
		ReportDir string `yaml:"report_dir"`
		JsonFormat bool  `yaml:"json_format"`
	} `yaml:"output"`
}

type AnalysisResult struct {
	Timestamp       time.Time  `json:"timestamp"`
	FileName        string     `json:"file_name"`
	FileType        string     `json:"file_type"`
	FileSize        int64      `json:"file_size"`
	ChunksProcessed int        `json:"chunks_processed"`
	Findings        Findings   `json:"findings"`
	RawAnalysis     string     `json:"raw_analysis,omitempty"`
}

type Findings struct {
	Credentials   []Credential   `json:"credentials,omitempty"`
	CVESignatures []CVESignature `json:"cve_signatures,omitempty"`
	Anomalies     []Anomaly      `json:"anomalies,omitempty"`
	IOCs          []IOC          `json:"iocs,omitempty"`
	Summary       string         `json:"summary"`
	Severity      string         `json:"severity"`
}

type Credential struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Location string `json:"location"`
	Context  string `json:"context"`
}

type CVESignature struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Confidence  string `json:"confidence"`
	Location    string `json:"location"`
}

type Anomaly struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Line        int    `json:"line,omitempty"`
	Context     string `json:"context"`
}

type IOC struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Source   string `json:"source"`
	Category string `json:"category"`
}

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

var cfg Config

func main() {
	var inputFile string
	var outputFile string
	var noColor bool

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		case "-o", "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--no-color":
			noColor = true
		default:
			if !strings.HasPrefix(args[i], "-") {
				inputFile = args[i]
			}
		}
	}

	if inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: No input file specified\n\n")
		printHelp()
		os.Exit(1)
	}

	if err := loadConfig("config.yaml"); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if noColor {
		disableColors()
	}

	result, err := analyzeFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[-] Analysis error: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	printResults(result, outputFile)

	if outputFile != "" {
		saveReport(result, outputFile)
	}
}

func printHelp() {
	fmt.Println("Usage: ai-analyze [OPTIONS] <file>")
	fmt.Println("")
	fmt.Println("Analyze security artifacts using AI.")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -h, --help          Show this help message")
	fmt.Println("  -o, --output FILE   Save report to file")
	fmt.Println("  --no-color          Disable colored output")
	fmt.Println("")
	fmt.Println("Supported file types: .log, .pcap, .json, .txt, .xml")
}

func disableColors() {
	ColorReset = ""
	ColorRed = ""
	ColorGreen = ""
	ColorYellow = ""
	ColorBlue = ""
	ColorCyan = ""
	ColorBold = ""
}

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		cfg = getDefaultConfig()
		return nil
	}
	return yaml.Unmarshal(data, &cfg)
}

func getDefaultConfig() Config {
	var c Config
	c.AiCore.Endpoint = "http://localhost:8080/analyze"
	c.AiCore.Model = "default"
	c.AiCore.Temperature = 0.3
	c.AiCore.MaxTokens = 4000
	c.AiCore.Timeout = 120
	c.Analysis.ChunkSize = MaxChunkSize
	c.Analysis.MaxFileSize = 100 * 1024 * 1024
	c.Output.ReportDir = "./reports"
	c.Output.JsonFormat = false
	return c
}

func analyzeFile(filePath string) (*AnalysisResult, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	if cfg.Analysis.MaxFileSize > 0 && fileInfo.Size() > cfg.Analysis.MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", fileInfo.Size(), cfg.Analysis.MaxFileSize)
	}

	fileType := detectFileType(filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	result := &AnalysisResult{
		Timestamp:       time.Now(),
		FileName:        filepath.Base(filePath),
		FileType:        fileType,
		FileSize:        fileInfo.Size(),
		Findings:        Findings{},
	}

	chunkSize := cfg.Analysis.ChunkSize
	if chunkSize <= 0 {
		chunkSize = MaxChunkSize
	}

	if len(content) <= chunkSize {
		analysis, err := sendToAiCore(content, fileType, filePath)
		if err != nil {
			return nil, err
		}
		result.ChunksProcessed = 1
		result.RawAnalysis = analysis
		result.Findings = parseAiResponse(analysis)
	} else {
		chunks := chunkContent(content, chunkSize)
		result.ChunksProcessed = len(chunks)

		var allFindings Findings
		allFindings.Summary = ""

		for i, chunk := range chunks {
			chunkLabel := fmt.Sprintf("Part %d/%d", i+1, len(chunks))
			analysis, err := sendToAiCore(chunk, fileType, fmt.Sprintf("%s [%s]", filePath, chunkLabel))
			if err != nil {
				continue
			}
			chunkFindings := parseAiResponse(analysis)
			mergeFindings(&allFindings, &chunkFindings)
			if allFindings.Summary != "" {
				allFindings.Summary += "\n\n"
			}
			allFindings.Summary += fmt.Sprintf("[%s]\n%s", chunkLabel, chunkFindings.Summary)
		}

		result.Findings = allFindings
		consolidated, err := sendConsolidation(&result.Findings)
		if err == nil {
			result.Findings = parseAiResponse(consolidated)
		}
	}

	result.Findings.Severity = assessSeverity(&result.Findings)

	return result, nil
}

func detectFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	magicBytes := map[string][]byte{
		".pcap":    {0xD4, 0xC3, 0xB2, 0xA1},
		".pcapng":  {0x0A, 0x0D, 0x0D, 0x0A},
	}

	data, err := os.ReadFile(filePath)
	if err == nil && len(data) >= 4 {
		for extension, magic := range magicBytes {
			if len(data) >= len(magic) {
				match := true
				for i, b := range magic {
					if data[i] != b {
						match = false
						break
					}
				}
				if match {
					return extension
				}
			}
		}
	}

	switch ext {
	case ".log":
		return "log"
	case ".pcap", ".pcapng":
		return "pcap"
	case ".json":
		return "json"
	case ".txt":
		return "text"
	case ".xml":
		return "xml"
	default:
		return "unknown"
	}
}

func chunkContent(content []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	lines := bytes.Split(content, []byte("\n"))

	var currentChunk []byte
	currentSize := 0

	for _, line := range lines {
		lineSize := len(line) + 1
		if currentSize+lineSize > chunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = nil
			currentSize = 0
		}
		currentChunk = append(currentChunk, line...)
		currentChunk = append(currentChunk, '\n')
		currentSize += lineSize
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

func sendToAiCore(content []byte, fileType, filePath string) (string, error) {
	jsonData, _ := json.Marshal(map[string]interface{}{
		"model":       cfg.AiCore.Model,
		"temperature": cfg.AiCore.Temperature,
		"max_tokens":  cfg.AiCore.MaxTokens,
		"system":      getSystemPrompt(fileType),
		"prompt":      fmt.Sprintf("Analyze this %s file: %s\n\nContent:\n%s", fileType, filePath, string(content)),
	})

	if strings.HasPrefix(cfg.AiCore.Endpoint, "http") {
		req, err := http.NewRequest("POST", cfg.AiCore.Endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return callAiCoreCli(content, fileType, filePath)
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: time.Duration(cfg.AiCore.Timeout) * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return callAiCoreCli(content, fileType, filePath)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		if resp.StatusCode != 200 {
			return callAiCoreCli(content, fileType, filePath)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return "", err
		}

		if response, ok := result["response"].(string); ok {
			return response, nil
		}
		if response, ok := result["content"].(string); ok {
			return response, nil
		}
		if response, ok := result["text"].(string); ok {
			return response, nil
		}

		return string(body), nil
	}

	return callAiCoreCli(content, fileType, filePath)
}

func callAiCoreCli(content []byte, fileType, filePath string) (string, error) {
	args := []string{
		"--model", cfg.AiCore.Model,
		"--temperature", fmt.Sprintf("%f", cfg.AiCore.Temperature),
		"--max-tokens", fmt.Sprintf("%d", cfg.AiCore.MaxTokens),
		"--system", getSystemPrompt(fileType),
		"--prompt", fmt.Sprintf("Analyze this %s file: %s\n\nContent:\n%s", fileType, filePath, string(content)),
	}

	cmd := exec.Command("ai-core", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		if out.Len() > 0 {
			return out.String(), nil
		}
		return "", fmt.Errorf("ai-core command failed: %w", err)
	}

	return out.String(), nil
}

func getSystemPrompt(fileType string) string {
	return fmt.Sprintf(`You are a security analyst specializing in analyzing log files, network captures (pcap), and security artifacts. Your task is to identify security-relevant information and provide a detailed analysis.

For the given file, identify and categorize the following:

1. CREDENTIALS:
   - API keys, tokens, secrets
   - Passwords (plaintext or encoded)
   - AWS keys, SSH keys, GPG keys
   - Authentication tokens, session IDs
   - JWT tokens
   - Database connection strings

2. CVE SIGNATURES:
   - Known CVE references (CVE-YYYY-NNNNN)
   - Vulnerability patterns (SQL injection, XSS, RCE indicators)
   - Exploit attempts
   - Patch level indicators

3. ANOMALIES:
   - Unusual access patterns
   - Failed authentication attempts
   - Suspicious commands or activities
   - Privilege escalation attempts
   - Unusual network behavior
   - Error patterns indicating compromise

4. INDICATORS OF COMPROMISE (IOCs):
   - IP addresses (especially suspicious/known malicious)
   - Domain names
   - File hashes (MD5, SHA1, SHA256)
   - URLs
   - Email addresses
   - Registry keys
   - Process names
   - Mutex names

Output format:
Provide your findings in a structured format with clear sections. Use severity ratings (Critical, High, Medium, Low, Info) where applicable.

For each finding, include:
- Type
- Value/Content
- Location (line number, timestamp, etc.)
- Context/Explanation
- Severity

File type: %s

Analyze thoroughly and provide actionable security insights.`, fileType)
}

func parseAiResponse(response string) Findings {
	var findings Findings

	lines := strings.Split(response, "\n")
	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		lower := strings.ToLower(line)

		if strings.Contains(lower, "credential") || strings.Contains(lower, "password") ||
			strings.Contains(lower, "api key") || strings.Contains(lower, "token") ||
			strings.Contains(lower, "secret") {
			currentSection = "credentials"
			continue
		}
		if strings.Contains(lower, "cve-") || strings.Contains(lower, "vulnerability") ||
			strings.Contains(lower, "exploit") {
			currentSection = "cves"
			continue
		}
		if strings.Contains(lower, "anomaly") || strings.Contains(lower, "suspicious") ||
			strings.Contains(lower, "unusual") || strings.Contains(lower, "failed") {
			currentSection = "anomalies"
			continue
		}
		if strings.Contains(lower, "ioc") || strings.Contains(lower, "indicator") ||
			strings.Contains(lower, "ip address") || strings.Contains(lower, "domain") ||
			strings.Contains(lower, "hash") || strings.Contains(lower, "url") {
			currentSection = "iocs"
			continue
		}
		if strings.Contains(lower, "summary") || strings.Contains(lower, "conclusion") {
			currentSection = "summary"
			continue
		}

		if currentSection == "summary" && line != "" {
			findings.Summary += line + "\n"
			continue
		}

		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") ||
			strings.HasPrefix(line, "•") {
			cleaned := strings.TrimLeft(line, "-*• ")
			if cleaned != "" {
				processFinding(&findings, currentSection, cleaned)
			}
		}
	}

	if findings.Summary == "" {
		findings.Summary = response
	}

	return findings
}

func processFinding(findings *Findings, section, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}

	switch section {
	case "credentials":
		findings.Credentials = append(findings.Credentials, Credential{
			Value:   content,
			Type:    "detected",
			Context: "Found in analysis",
		})
	case "cves":
		if strings.Contains(content, "CVE-") {
			parts := strings.Split(content, " ")
			for _, part := range parts {
				if strings.Contains(part, "CVE-") {
					findings.CVESignatures = append(findings.CVESignatures, CVESignature{
						ID:          part,
						Description: content,
						Confidence:  "medium",
					})
					break
				}
			}
		}
	case "anomalies":
		findings.Anomalies = append(findings.Anomalies, Anomaly{
			Description: content,
			Type:        "detected",
			Context:     "Found in analysis",
		})
	case "iocs":
		if isIOC(content) {
			findings.IOCs = append(findings.IOCs, IOC{
				Value:    content,
				Type:     "detected",
				Category: "network",
			})
		}
	}
}

func isIOC(s string) bool {
	s = strings.TrimSpace(s)

	ipPattern := `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
	if matched, _ := regexp.MatchString(ipPattern, s); matched {
		return true
	}

	domainPattern := `^[a-zA-Z0-9][-a-zA-Z0-9]*\.[a-zA-Z]{2,}`
	if matched, _ := regexp.MatchString(domainPattern, s); matched {
		return true
	}

	hashPatterns := []string{`^[a-fA-F0-9]{32}$`, `^[a-fA-F0-9]{40}$`, `^[a-fA-F0-9]{64}$`}
	for _, p := range hashPatterns {
		if matched, _ := regexp.MatchString(p, s); matched {
			return true
		}
	}

	urlPattern := `^https?://`
	if matched, _ := regexp.MatchString(urlPattern, s); matched {
		return true
	}

	return false
}

func mergeFindings(target, source *Findings) {
	target.Credentials = append(target.Credentials, source.Credentials...)
	target.CVESignatures = append(target.CVESignatures, source.CVESignatures...)
	target.Anomalies = append(target.Anomalies, source.Anomalies...)
	target.IOCs = append(target.IOCs, source.IOCs...)

	if source.Summary != "" {
		if target.Summary != "" {
			target.Summary += "\n"
		}
		target.Summary += source.Summary
	}
}

func sendConsolidation(findings *Findings) (string, error) {
	summaryJSON, _ := json.Marshal(findings)

	prompt := fmt.Sprintf(`Consolidate and summarize the following security findings into a coherent report. Identify the most critical issues and provide a clear summary:

Findings: %s`, string(summaryJSON))

	jsonData, _ := json.Marshal(map[string]interface{}{
		"model":       cfg.AiCore.Model,
		"temperature": 0.3,
		"max_tokens":  2000,
		"system":      "You are a security analyst. Consolidate and summarize findings.",
		"prompt":      prompt,
	})

	if strings.HasPrefix(cfg.AiCore.Endpoint, "http") {
		req, _ := http.NewRequest("POST", cfg.AiCore.Endpoint, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: time.Duration(cfg.AiCore.Timeout) * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return string(body), nil
	}

	return callAiCoreCli([]byte(prompt), "summary", "consolidated")
}

func assessSeverity(findings *Findings) string {
	severityScore := 0

	for _, cred := range findings.Credentials {
		if strings.Contains(strings.ToLower(cred.Type), "password") ||
			strings.Contains(strings.ToLower(cred.Type), "key") ||
			strings.Contains(strings.ToLower(cred.Type), "secret") {
			severityScore += 3
		} else {
			severityScore += 2
		}
	}

	for range findings.CVESignatures {
		severityScore += 3
	}

	for range findings.Anomalies {
		severityScore += 1
	}

	for range findings.IOCs {
		severityScore += 2
	}

	if severityScore >= 10 {
		return "Critical"
	}
	if severityScore >= 6 {
		return "High"
	}
	if severityScore >= 3 {
		return "Medium"
	}
	if severityScore >= 1 {
		return "Low"
	}
	return "Info"
}

func printResults(result *AnalysisResult, outputFile string) {
	fmt.Printf("%s[%s] AI Security Analysis%s\n", ColorBold, result.Timestamp.Format("2006-01-02 15:04:05"), ColorReset)
	fmt.Printf("%s[*] File: %s%s\n", ColorCyan, result.FileName, ColorReset)
	fmt.Printf("%s[*] Type: %s | Size: %s | Chunks: %d%s\n", ColorCyan, result.FileType, formatFileSize(result.FileSize), result.ChunksProcessed, ColorReset)

	fmt.Printf("\n%s[%s] SEVERITY: %s%s\n", ColorBold, ColorRed, result.Findings.Severity, ColorReset)

	if len(result.Findings.Credentials) > 0 {
		fmt.Printf("\n%s[!] Credentials Found (%d)%s\n", ColorRed, len(result.Findings.Credentials), ColorReset)
		for _, cred := range result.Findings.Credentials {
			fmt.Printf("  %s[!]%s %s\n", ColorRed, ColorReset, cred.Value)
		}
	}

	if len(result.Findings.CVESignatures) > 0 {
		fmt.Printf("\n%s[!] CVE Signatures (%d)%s\n", ColorRed, len(result.Findings.CVESignatures), ColorReset)
		for _, cve := range result.Findings.CVESignatures {
			fmt.Printf("  %s[*]%s %s - %s\n", ColorYellow, ColorReset, cve.ID, cve.Description)
		}
	}

	if len(result.Findings.Anomalies) > 0 {
		fmt.Printf("\n%s[*] Anomalies (%d)%s\n", ColorYellow, len(result.Findings.Anomalies), ColorReset)
		for _, anon := range result.Findings.Anomalies {
			fmt.Printf("  %s- %s%s\n", ColorYellow, anon.Description, ColorReset)
		}
	}

	if len(result.Findings.IOCs) > 0 {
		fmt.Printf("\n%s[*] Indicators of Compromise (%d)%s\n", ColorYellow, len(result.Findings.IOCs), ColorReset)
		for _, ioc := range result.Findings.IOCs {
			fmt.Printf("  %s[*]%s [%s] %s\n", ColorYellow, ColorReset, ioc.Type, ioc.Value)
		}
	}

	fmt.Printf("\n%s[*] Summary:%s\n", ColorCyan, ColorReset)
	fmt.Println(result.Findings.Summary)

	if outputFile != "" {
		fmt.Printf("\n%s[+] Report saved to: %s%s\n", ColorGreen, outputFile, ColorReset)
	}
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func saveReport(result *AnalysisResult, outputFile string) {
	dir := filepath.Dir(outputFile)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s[!] Error saving report: %v%s\n", ColorRed, err, ColorReset)
		return
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "%s[!] Error saving report: %v%s\n", ColorRed, err, ColorReset)
	}
}
