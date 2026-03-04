package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Tools struct {
		Nmap    string `yaml:"nmap"`
		Dirb    string `yaml:"dirb"`
		DnsEnum string `yaml:"dnsenum"`
		AiCore  string `yaml:"ai_core"`
	} `yaml:"tools"`
	Nmap struct {
		QuickScan   string `yaml:"quick_scan"`
		FullScan    string `yaml:"full_scan"`
		OutputFormat string `yaml:"output_format"`
	} `yaml:"nmap"`
	Dirb struct {
		Wordlist     string `yaml:"wordlist"`
		OutputFormat string `yaml:"output_format"`
	} `yaml:"dirb"`
	DnsEnum struct {
		Threads     int    `yaml:"threads"`
		OutputFormat string `yaml:"output_format"`
	} `yaml:"dnsenum"`
	Ai struct {
		Model      string  `yaml:"model"`
		Temperature float64 `yaml:"temperature"`
		MaxTokens  int     `yaml:"max_tokens"`
	} `yaml:"ai"`
	Output struct {
		BaseDir      string `yaml:"base_dir"`
		KeepRaw      bool   `yaml:"keep_raw"`
		SummaryFormat string `yaml:"summary_format"`
	} `yaml:"output"`
}

type NmapResult struct {
	Host    string   `json:"host"`
	Ports   []Port   `json:"ports"`
	OS      string   `json:"os"`
	Scripts []Script `json:"scripts"`
}

type Port struct {
	Protocol string `json:"protocol"`
	PortID   int    `json:"port_id"`
	State    string `json:"state"`
	Service  string `json:"service"`
	Version  string `json:"version"`
}

type Script struct {
	Name string `json:"name"`
	Output string `json:"output"`
}

type DirbResult struct {
	URL      string   `json:"url"`
	Code     int      `json:"code"`
	Size     int64    `json:"size"`
	Path     string   `json:"path"`
}

type DnsResult struct {
	Domain   string   `xml:"domain"`
	Records  []Record `xml:"record"`
}

type Record struct {
	Type  string `xml:"type"`
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type ReconResult struct {
	Target      string     `json:"target"`
	Timestamp   time.Time  `json:"timestamp"`
	NmapResults []NmapResult `json:"nmap_results"`
	DirbResults []DirbResult `json:"dirb_results"`
	DnsResults  []DnsResult  `json:"dns_results"`
	RawOutputs  map[string]string `json:"raw_outputs"`
}

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

var cfg Config
var outputDir string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: auto-recon <target>")
		fmt.Println("Target: IP address or domain name")
		os.Exit(1)
	}

	target := os.Args[1]
	
	if err := loadConfig("config.yaml"); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	outputDir = filepath.Join(cfg.Output.BaseDir, fmt.Sprintf("%s_%s", target, time.Now().Format("20060102_150405")))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s[+] Smart Recon Pipeline Started for: %s%s\n", ColorCyan, target, ColorReset)
	fmt.Printf("[*] Output directory: %s\n", outputDir)

	result := &ReconResult{
		Target:     target,
		Timestamp:  time.Now(),
		RawOutputs: make(map[string]string),
	}

	fmt.Printf("\n%s[*] Phase 1: Nmap Quick Scan%s\n", ColorBlue, ColorReset)
	nmapOutput, err := runNmap(target)
	if err != nil {
		fmt.Printf("%s[-] Nmap error: %v%s\n", ColorRed, err, ColorReset)
	} else {
		result.NmapResults = parseNmapOutput(nmapOutput)
		result.RawOutputs["nmap"] = nmapOutput
		saveOutput("nmap_results.json", nmapOutput)
	}

	fmt.Printf("\n%s[*] Phase 2: Directory Enumeration (dirb)%s\n", ColorBlue, ColorReset)
	dirbOutput, err := runDirb(target)
	if err != nil {
		fmt.Printf("%s[-] Dirb error: %v%s\n", ColorRed, err, ColorReset)
	} else {
		result.DirbResults = parseDirbOutput(dirbOutput)
		result.RawOutputs["dirb"] = dirbOutput
		saveOutput("dirb_results.json", dirbOutput)
	}

	fmt.Printf("\n%s[*] Phase 3: DNS Enumeration%s\n", ColorBlue, ColorReset)
	dnsOutput, err := runDnsEnum(target)
	if err != nil {
		fmt.Printf("%s[-] DNS enum error: %v%s\n", ColorRed, err, ColorReset)
	} else {
		result.DnsResults = parseDnsOutput(dnsOutput)
		result.RawOutputs["dns"] = dnsOutput
		saveOutput("dns_results.xml", dnsOutput)
	}

	fmt.Printf("\n%s[*] Phase 4: AI Analysis & Summary%s\n", ColorBlue, ColorReset)
	summary, err := runAiAnalysis(result)
	if err != nil {
		fmt.Printf("%s[-] AI analysis error: %v%s\n", ColorRed, err, ColorReset)
		printLocalSummary(result)
	} else {
		printAiSummary(summary)
	}

	saveOutput("recon_full_report.json", marshalJSON(result))
	
	fmt.Printf("\n%s[+] Reconnaissance complete!%s\n", ColorGreen, ColorReset)
	fmt.Printf("[*] Full report: %s\n", filepath.Join(outputDir, "recon_full_report.json"))
}

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &cfg)
}

func runNmap(target string) (string, error) {
	args := strings.Fields(cfg.Nmap.QuickScan)
	args = append(args, "-oX", filepath.Join(outputDir, "nmap.xml"))
	args = append(args, target)
	
	cmd := exec.Command(cfg.Tools.Nmap, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	if err := cmd.Run(); err != nil {
		return "", err
	}
	
	xmlData, err := os.ReadFile(filepath.Join(outputDir, "nmap.xml"))
	if err != nil {
		return "", err
	}
	
	return string(xmlData), nil
}

func runDirb(target string) (string, error) {
	url := target
	if !strings.HasPrefix(target, "http") {
		url = "http://" + target
	}
	
	args := []string{"-o", filepath.Join(outputDir, "dirb.txt"), url, cfg.Dirb.Wordlist}
	
	cmd := exec.Command(cfg.Tools.Dirb, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	cmd.Run()
	
	data, err := os.ReadFile(filepath.Join(outputDir, "dirb.txt"))
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

func runDnsEnum(target string) (string, error) {
	args := []string{"-f", "false", "-o", filepath.Join(outputDir, "dns.xml"), target}
	
	cmd := exec.Command(cfg.Tools.DnsEnum, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	cmd.Run()
	
	data, err := os.ReadFile(filepath.Join(outputDir, "dns.xml"))
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

func runAiAnalysis(result *ReconResult) (string, error) {
	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`Analyze the following reconnaissance results and provide a security summary. Identify:
1. Open ports and services
2. Potential vulnerabilities
3. Interesting directories/files
4. DNS findings
5. Recommendations

Results: %s`, string(jsonData))

	args := []string{
		"--model", cfg.Ai.Model,
		"--temperature", fmt.Sprintf("%f", cfg.Ai.Temperature),
		"--max-tokens", fmt.Sprintf("%d", cfg.Ai.MaxTokens),
		"--prompt", prompt,
	}

	cmd := exec.Command(cfg.Tools.AiCore, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}

func parseNmapOutput(xmlData string) []NmapResult {
	var results []NmapResult
	
	type NmapRun struct {
		Host []struct {
			Address string `xml:"address>addr"`
			Ports   struct {
				Port []struct {
					Protocol string `xml:"protocol"`
					PortID   int    `xml:"portid"`
					State    string `xml:"state>state"`
					Service  struct {
						Name    string `xml:"name"`
						Product string `xml:"product"`
						Version string `xml:"version"`
					} `xml:"service"`
				} `xml:"port"`
			} `xml:"ports"`
		} `xml:"host"`
	}

	var run NmapRun
	if err := xml.Unmarshal([]byte(xmlData), &run); err != nil {
		return results
	}

	for _, host := range run.Host {
		result := NmapResult{
			Host: host.Address,
		}
		for _, port := range host.Ports.Port {
			result.Ports = append(result.Ports, Port{
				Protocol: port.Protocol,
				PortID:   port.PortID,
				State:    port.State,
				Service:  port.Service.Name,
				Version:  port.Service.Product + " " + port.Service.Version,
			})
		}
		results = append(results, result)
	}

	return results
}

func parseDirbOutput(text string) []DirbResult {
	var results []DirbResult
	lines := strings.Split(text, "\n")
	
	for _, line := range lines {
		if strings.Contains(line, "+") && strings.Contains(line, "(CODE:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				path := strings.TrimSuffix(strings.TrimPrefix(parts[0], "+"), "/")
				codeIdx := strings.Index(line, "(CODE:")
				if codeIdx > 0 {
					codeStr := strings.TrimSuffix(strings.TrimPrefix(line[codeIdx:], "(CODE:"), ")")
					var code int
					fmt.Sscanf(codeStr, "%d", &code)
					results = append(results, DirbResult{
						URL:  path,
						Code: code,
					})
				}
			}
		}
	}
	
	return results
}

func parseDnsOutput(xmlData string) []DnsResult {
	var results []DnsResult
	
	type DnsEnumRun struct {
		Domain string `xml:"domain"`
		Host   []struct {
			IP string `xml:"ip"`
		} `xml:"host"`
	}

	var run DnsEnumRun
	if err := xml.Unmarshal([]byte(xmlData), &run); err != nil {
		return results
	}

	result := DnsResult{Domain: run.Domain}
	for _, host := range run.Host {
		result.Records = append(result.Records, Record{
			Type:  "A",
			Value: host.IP,
		})
	}
	results = append(results, result)

	return results
}

func printLocalSummary(result *ReconResult) {
	fmt.Printf("\n%s=== Local Summary ===%s\n", ColorYellow, ColorReset)
	
	fmt.Printf("\n%s[*] Open Ports & Services:%s\n", ColorCyan, ColorReset)
	for _, nmapRes := range result.NmapResults {
		for _, port := range nmapRes.Ports {
			if port.State == "open" {
				fmt.Printf("  [%s%d%s] %s - %s %s\n", 
					ColorGreen, port.PortID, ColorReset, 
					port.Service, ColorYellow, port.Version, ColorReset)
			}
		}
	}
	
	fmt.Printf("\n%s[*] Discovered Directories:%s\n", ColorCyan, ColorReset)
	for _, dirbRes := range result.DirbResults {
		fmt.Printf("  [%s%d%s] %s\n", ColorGreen, dirbRes.Code, ColorReset, dirbRes.URL)
	}
	
	fmt.Printf("\n%s[*] DNS Records:%s\n", ColorCyan, ColorReset)
	for _, dnsRes := range result.DnsResults {
		for _, record := range dnsRes.Records {
			fmt.Printf("  [%s] %s -> %s\n", ColorGreen, record.Type, record.Value)
		}
	}
}

func printAiSummary(summary string) {
	fmt.Printf("\n%s=== AI Security Analysis ===%s\n", ColorYellow, ColorReset)
	fmt.Println(summary)
}

func saveOutput(filename, data string) {
	path := filepath.Join(outputDir, filename)
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		fmt.Printf("%s[!] Error saving %s: %v%s\n", ColorRed, filename, err, ColorReset)
	}
}

func marshalJSON(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
