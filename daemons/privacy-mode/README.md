# Privacy Mode Daemon

A lightweight system daemon for OffTopic Linux that provides configurable privacy protection features for user workstations.

## Features

### MAC Address Randomization
Randomizes network interface MAC addresses to prevent hardware fingerprinting. When enabled with `randomize_on_network_up`, the daemon automatically generates a new random MAC address each time a network interface comes up.

### TTL Hardening
Sets the default IP Time-To-Live to a generic value, making network traffic analysis more difficult and preventing easy OS fingerprinting through TTL values.

### Temporary Files Cleanup
Automatically cleans up files in `/tmp` older than 24 hours during logout, helping maintain a cleaner system and reduce temporary data footprint.

## Installation

1. Build the daemon:
```bash
go build -o privacy-mode-daemon main.go
```

2. Install the binary:
```bash
sudo cp privacy-mode-daemon /usr/bin/
```

3. Install the configuration:
```bash
sudo mkdir -p /etc/privacy-mode
sudo cp config.json /etc/privacy-mode/config.json
```

4. Install the systemd service:
```bash
sudo cp privacy-mode-daemon.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable privacy-mode-daemon
sudo systemctl start privacy-mode-daemon
```

## Configuration

Edit `/etc/privacy-mode/config.json`:

| Setting | Type | Description |
|---------|------|-------------|
| `enable_mac_randomization` | boolean | Enable MAC address randomization |
| `randomize_on_network_up` | boolean | Randomize MAC when network interfaces come up |
| `enable_ttl_hardening` | boolean | Enable TTL hardening |
| `ttl_value` | integer | TTL value to set (default: 64) |
| `enable_tmp_cleanup` | boolean | Enable /tmp cleanup on logout |
| `debug` | boolean | Enable debug logging |

## Requirements

- Go 1.21+
- Linux kernel
- iproute2 (for MAC randomization)
- systemd

## Usage

### Enable MAC Randomization
```json
{
  "enable_mac_randomization": true,
  "randomize_on_network_up": true
}
```

### Enable TTL Hardening
```json
{
  "enable_ttl_hardening": true,
  "ttl_value": 64
}
```

### Enable TMP Cleanup
```json
{
  "enable_tmp_cleanup": true
}
```

## Security Considerations

This daemon is designed for legitimate privacy protection on user-controlled systems. All features are:
- Configurable via the config file
- Non-destructive (MAC changes are temporary until next reboot unless persistent)
- Compliant with standard Linux networking practices

## License

MIT License
