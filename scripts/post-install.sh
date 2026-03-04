#!/bin/bash
# Kali Openbox - Post-installation script (PopOS! style)
# Run: sudo ./post-install.sh

set -e

echo "==================================="
echo "  Kali Openbox Setup"
echo "  PopOS! inspired - Lean & Clean"
echo "==================================="

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root: sudo $0"
    exit 1
fi

echo "[1/7] Updating system..."
apt update && apt upgrade -y

echo "[2/7] Installing Openbox & Desktop..."
apt install -y \
    openbox \
    obconf \
    obmenu \
    obmenu-generator \
    rofi \
    kitty \
    thunar \
    thunar-archive-plugin \
    thunar-volman \
    nitrogen \
    picom \
    gtk-theme-config \
    xfce4-notifyd \
    xfce4-screenshooter \
    policykit-1-gnome

echo "[3/7] Installing Bluetooth & WiFi..."
apt install -y \
    bluez \
    bluez-tools \
    blueman \
    network-manager \
    network-manager-gnome \
    network-manager-openvpn \
    network-manager-pptp \
    network-manager-vpnc \
    wireless-tools \
    wpagui \
    iwd

echo "[4/7] Installing Nvidia drivers..."
apt install -y \
    nvidia-driver \
    nvidia-settings \
    nvidia-smi

echo "[5/7] Installing utilities..."
apt install -y \
    git \
    curl \
    wget \
    zsh \
    htop \
    neofetch \
    vim \
    nano \
    tree \
    tmux \
    pulseaudio \
    pavucontrol \
    volumeicon-apt \
    xclip \
    xsel \
    xdotool \
    xlock

# Setup Openbox config
echo "[6/7] Setting up configuration..."
mkdir -p ~/.config/openbox
mkdir -p ~/.config/nitrogen
mkdir -p ~/.config/picom
mkdir -p ~/.local/share/wallpapers

# Download wallpaper
wget -O ~/.local/share/wallpapers/kali-dark.jpg \
    https://wallpapercave.com/wp/wp6908708.jpg 2>/dev/null || true

# Set default editor
export EDITOR=nano
export VISUAL=nano

# Set zsh as default
chsh -s /bin/zsh

echo "[7/7] Setup complete!"

echo ""
echo "==================================="
echo "  QUICK START"
echo "==================================="
echo ""
echo "  Press SUPER key for Spotlight search"
echo "  Alt+F2 for Run dialog"
echo "  Ctrl+Return for Terminal (Kitty)"
echo "  Print Screen for Screenshot"
echo "  Ctrl+Alt+L to Lock"
echo ""
echo "  Logout and select 'Openbox'"
echo "==================================="
