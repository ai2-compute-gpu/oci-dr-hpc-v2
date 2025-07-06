#!/bin/bash
# Install shapes.json file to system location

set -e

SHAPES_SRC="internal/shapes/shapes.json"
SHAPES_DEST="/etc/oci-dr-hpc-shapes.json"

echo "Installing shapes configuration file..."

# Check if source file exists
if [ ! -f "$SHAPES_SRC" ]; then
    echo "Error: Source file $SHAPES_SRC not found"
    exit 1
fi

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root (use sudo)"
    exit 1
fi

# Copy the file
cp "$SHAPES_SRC" "$SHAPES_DEST"

# Set appropriate permissions
chmod 644 "$SHAPES_DEST"

echo "Successfully installed shapes configuration to $SHAPES_DEST"
echo "The diagnostic tool will now use the system-wide shapes configuration." 