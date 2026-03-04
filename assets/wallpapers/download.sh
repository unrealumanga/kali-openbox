#!/bin/bash
# Download default wallpapers for OffTopic Linux
# Run during build process

WALLPAPER_DIR="/usr/share/backgrounds/offtopic"
mkdir -p "$WALLPAPER_DIR"

echo "Downloading OffTopic wallpapers..."

# Download dark wallpaper (Kali-style dark)
wget -O "$WALLPAPER_DIR/offtopic-dark.jpg" \
    "https://images.unsplash.com/photo-1557683316-973673baf926?w=1920" \
    2>/dev/null || true

# Download minimal wallpaper
wget -O "$WALLPAPER_DIR/offtopic-minimal.jpg" \
    "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=1920" \
    2>/dev/null || true

# Download cyber/neon style
wget -O "$WALLPAPER_DIR/offtopic-neon.jpg" \
    "https://images.unsplash.com/photo-1550684848-fac1c5b4e853?w=1920" \
    2>/dev/null || true

# Set default wallpaper
if [ -f "$WALLPAPER_DIR/offtopic-dark.jpg" ]; then
    echo "$WALLPAPER_DIR/offtopic-dark.jpg" > /etc/default/desktop-background
fi

echo "Wallpapers installed!"
