# OffTopic Linux

Lightweight Kali Linux-based OS with Openbox desktop environment and native Nvidia support.

## Features

- **Spotlight Search**: Press `Super` for instant app/file search (PopOS! style)
- **Lean & Clean**: No panels, no clutter - just you and your tools
- **Nvidia Ready**: Native Nvidia driver and CUDA support built-in
- **Bluetooth & WiFi**: Ready out of the box
- **Pentester-focused**: Essential security tools included

## Quick Start

### Download ISO
Pre-built ISOs available in [Releases](https://github.com/yourusername/kali-openbox/releases)

### Build from Source

```bash
git clone https://github.com/yourusername/kali-openbox.git
cd kali-openbox
sudo ./build.sh
```

## Installation (Easy!)

### Option 1: Install from ISO (Beginner Friendly)
1. **Download** the ISO from releases
2. **Flash** to USB with [Rufus](https://rufus.ie/) or [Etcher](https://etcher.balena.io/)
3. **Boot** from USB
4. **Double-click** "Install OffTopic Linux" on desktop
5. **Follow** the guided steps (like Windows/Mac install!)

### Option 2: Run Live (Try before install)
1. Boot from USB
2. Select "Try OffTopic"
3. Test everything works
4. Click Install icon when ready

### Option 3: Post-Install Script (Advanced)
```bash
# On existing Kali/Debian
git clone https://github.com/yourusername/kali-openbox
cd kali-openbox
sudo ./scripts/post-install.sh
```

## Requirements

- 40GB+ disk space
- 4GB+ RAM
- Nvidia GPU (optional)

## ISO Size

- **~2.5GB** (similar to PopOS!, lighter than default Kali 4GB)

## Visual Design

### Desktop
- Dark minimalist theme (cyberpunk/green accent)
- Custom wallpapers (3 styles included)
- Spotlight search with Rofi
- Clean Openbox desktop

### Login Screen
- LightDM with custom dark theme
- Wallpaper background
- Papirus-Dark icons

### Boot
- Plymouth boot splash
- GRUB theme (optional)

## Credits

- [Kali Linux](https://www.kali.org/)
- [Openbox](http://openbox.org/)
- [Wallhaven](https://wallhaven.cc)

## License

MIT License
