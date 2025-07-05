#!/bin/bash
# setup-logging.sh - Setup log directory for oci-dr-hpc-v2
# This script should be run during package installation

LOG_DIR="/var/log/oci-dr-hpc"
LOG_FILE="$LOG_DIR/oci-dr-hpc.log"

echo "Setting up logging for oci-dr-hpc-v2..."

# Create log directory if it doesn't exist
if [ ! -d "$LOG_DIR" ]; then
    mkdir -p "$LOG_DIR"
    echo "Created log directory: $LOG_DIR"
fi

# Determine appropriate group for log files based on distribution
LOG_GROUP="adm"
if [ -f /etc/oracle-release ] || [ -f /etc/redhat-release ]; then
    # Oracle Linux/RHEL - check if adm group exists, otherwise use wheel
    if ! getent group adm >/dev/null 2>&1; then
        LOG_GROUP="wheel"
        echo "Using wheel group for Oracle Linux/RHEL compatibility"
    fi
elif [ -f /etc/debian_version ]; then
    # Debian/Ubuntu - use adm group (standard)
    LOG_GROUP="adm"
    echo "Using adm group for Debian/Ubuntu"
else
    # Fallback - check what's available
    if getent group adm >/dev/null 2>&1; then
        LOG_GROUP="adm"
    elif getent group wheel >/dev/null 2>&1; then
        LOG_GROUP="wheel"
    else
        LOG_GROUP="root"
        echo "Warning: Neither adm nor wheel group found, using root"
    fi
fi

# Set directory permissions to allow group write
chown root:$LOG_GROUP "$LOG_DIR"
chmod 775 "$LOG_DIR"
echo "Set directory permissions: $LOG_DIR (root:$LOG_GROUP 775)"

# Remove any existing directory with the log file name
if [ -d "$LOG_FILE" ]; then
    rm -rf "$LOG_FILE"
    echo "Removed existing directory: $LOG_FILE"
fi

# Create log file if it doesn't exist
if [ ! -f "$LOG_FILE" ]; then
    touch "$LOG_FILE"
    echo "Created log file: $LOG_FILE"
fi

# Set log file permissions
chown root:$LOG_GROUP "$LOG_FILE"
chmod 664 "$LOG_FILE"
echo "Set file permissions: $LOG_FILE (root:$LOG_GROUP 664)"

echo "Logging setup completed successfully!"
echo "Users in the '$LOG_GROUP' group can now write to the log file."