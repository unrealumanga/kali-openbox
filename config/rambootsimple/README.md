# RAM-Only Boot Configuration
# Add to your live-build config or GRUB

# In config/hooks/01-base.hook.chroot

# Create live-boot RAM overlay script
mkdir -p /etc/initramfs-tools/hooks
cat > /etc/initramfs-tools/hooks/ram-overlay << 'EOF'
#!/bin/sh
# Enable RAM overlay for live boot
PREREQ="overlayfs"
prereqs() { echo "$PREREQ"; }
case $1 in
    prereqs) prereqs; exit 0;;
esac

. /usr/share/initramfs-tools/hook-functions
copy_exec /sbin/mount.fuse /sbin
EOF
chmod +x /etc/initramfs-tools/hooks/ram-overlay

# GRUB boot parameters for RAM-only mode
# Add to /etc/default/grub:
# GRUB_CMDLINE_LIVE_DEFAULT="toram overlayfs=overlayfs"

# Create persistent storage detection script
cat > /usr/local/bin/live-detect.sh << 'EOF'
#!/bin/bash
# Detect if running in live RAM mode

if grep -q "toram\|overlayfs" /proc/cmdline 2>/dev/null; then
    echo "Running in RAM-only mode"
    exit 0
else
    echo "Running with disk access"
    exit 1
fi
EOF
chmod +x /usr/local/bin/live-detect.sh
