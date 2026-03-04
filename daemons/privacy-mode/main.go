package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/jochencem/notify"
)

type Config struct {
	EnableMACRandomization bool   `json:"enable_mac_randomization"`
	RandomizeOnNetworkUp   bool   `json:"randomize_on_network_up"`
	EnableTTLHardening     bool   `json:"enable_ttl_hardening"`
	TTLValue               int    `json:"ttl_value"`
	EnableTmpCleanup       bool   `json:"enable_tmp_cleanup"`
	Debug                  bool   `json:"debug"`
}

var config Config
var configPath = "/etc/privacy-mode/config.json"
var defaultTTL = 64

func loadConfig() error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &config)
}

func logDebug(msg string) {
	if config.Debug {
		log.Printf("[DEBUG] %s", msg)
	}
}

func randomizeMAC() error {
	logDebug("Starting MAC address randomization")

	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to get interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		mac, err := generateRandomMAC()
		if err != nil {
			log.Printf("Failed to generate MAC for %s: %v", iface.Name, err)
			continue
		}

		cmd := exec.Command("ip", "link", "set", iface.Name, "down")
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to bring down interface %s: %v", iface.Name, err)
			continue
		}

		cmd = exec.Command("ip", "link", "set", iface.Name, "address", mac)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to set MAC for %s: %v", iface.Name, err)
			cmd = exec.Command("ip", "link", "set", iface.Name, "up")
			cmd.Run()
			continue
		}

		cmd = exec.Command("ip", "link", "set", iface.Name, "up")
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to bring up interface %s: %v", iface.Name, err)
			continue
		}

		log.Printf("Randomized MAC address for %s to %s", iface.Name, mac)
	}

	return nil
}

func generateRandomMAC() (string, error) {
	buf := make([]byte, 6)
	_, err := randRead(buf)
	if err != nil {
		return "", err
	}

	buf[0] = (buf[0] & 0xfe) | 0x02

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]), nil
}

func randRead(b []byte) (int, error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Read(b)
}

func setTTL(ttl int) error {
	logDebug(fmt.Sprintf("Setting default TTL to %d", ttl))

	sysctlPath := "/proc/sys/net/ipv4/ip_default_ttl"
	if err := os.WriteFile(sysctlPath, []byte(fmt.Sprintf("%d", ttl)), 0644); err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	log.Printf("Set default TTL to %d", ttl)
	return nil
}

func cleanupTmp() error {
	logDebug("Cleaning up /tmp directory")

	tmpDir := "/tmp"
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read /tmp: %w", err)
	}

	cleaned := 0
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if time.Since(info.ModTime()) > 24*time.Hour {
			path := filepath.Join(tmpDir, entry.Name())
			if err := os.RemoveAll(path); err == nil {
				cleaned++
			}
		}
	}

	log.Printf("Cleaned %d old entries from /tmp", cleaned)
	return nil
}

func handleNetworkChange() {
	if config.EnableMACRandomization {
		if err := randomizeMAC(); err != nil {
			log.Printf("MAC randomization failed: %v", err)
		}
	}
}

func setupNetworkMonitoring() error {
	c, err := notify.New()
	if err != nil {
		return fmt.Errorf("failed to create notifier: %w", err)
	}

	if err := c.Watch("/sys/class/net"); err != nil {
		return fmt.Errorf("failed to watch network: %w", err)
	}

	go func() {
		for range c.Event {
			handleNetworkChange()
		}
	}()

	return nil
}

func handleLogout() {
	if config.EnableTmpCleanup {
		if err := cleanupTmp(); err != nil {
			log.Printf("TMP cleanup failed: %v", err)
		}
	}
}

func setupLogoutMonitoring() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)

	go func() {
		for range sigs {
			handleLogout()
		}
	}()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting privacy-mode daemon")

	if err := loadConfig(); err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		config = Config{
			EnableMACRandomization:  false,
			RandomizeOnNetworkUp:   false,
			EnableTTLHardening:      false,
			TTLValue:                64,
			EnableTmpCleanup:        false,
			Debug:                   false,
		}
	}

	logDebug(fmt.Sprintf("Config loaded: %+v", config))

	if config.EnableTTLHardening {
		if err := setTTL(config.TTLValue); err != nil {
			log.Printf("TTL hardening failed: %v", err)
		}
	}

	if config.EnableMACRandomization && config.RandomizeOnNetworkUp {
		if err := setupNetworkMonitoring(); err != nil {
			log.Printf("Network monitoring setup failed: %v", err)
		} else {
			log.Println("Network monitoring enabled")
		}
	}

	setupLogoutMonitoring()
	log.Println("Logout monitoring enabled")

	daemon.SdNotify(false, "READY=1")

	for {
		time.Sleep(24 * time.Hour)
	}
}
