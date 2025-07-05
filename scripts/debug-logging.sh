#!/bin/bash
# debug-logging.sh - Debug logging issues on Oracle Linux

echo "=== OCI DR HPC Logging Debug ==="
echo "Date: $(date)"
echo "User: $(whoami)"
echo "Groups: $(groups)"
echo

echo "=== System Information ==="
if [ -f /etc/oracle-release ]; then
    echo "Oracle Linux detected:"
    cat /etc/oracle-release
elif [ -f /etc/redhat-release ]; then
    echo "RHEL detected:"
    cat /etc/redhat-release
elif [ -f /etc/debian_version ]; then
    echo "Debian/Ubuntu detected:"
    cat /etc/debian_version
fi
echo

echo "=== Available Groups ==="
getent group | grep -E "^(adm|wheel):" || echo "Neither adm nor wheel group found"
echo

echo "=== Log Directory Status ==="
if [ -d /var/log/oci-dr-hpc ]; then
    echo "Log directory exists:"
    ls -la /var/log/oci-dr-hpc/
    echo
    echo "Directory permissions:"
    stat /var/log/oci-dr-hpc/
    echo
    
    if [ -f /var/log/oci-dr-hpc/oci-dr-hpc.log ]; then
        echo "Log file exists:"
        stat /var/log/oci-dr-hpc/oci-dr-hpc.log
    elif [ -d /var/log/oci-dr-hpc/oci-dr-hpc ]; then
        echo "ERROR: oci-dr-hpc exists as DIRECTORY instead of file!"
        ls -la /var/log/oci-dr-hpc/oci-dr-hpc/
    else
        echo "No log file or directory found"
    fi
else
    echo "Log directory does not exist"
fi
echo

echo "=== User Write Test ==="
TEST_FILE="/var/log/oci-dr-hpc/test-write-$(date +%s).tmp"
if echo "test" > "$TEST_FILE" 2>/dev/null; then
    echo "✅ User can write to log directory"
    rm -f "$TEST_FILE"
else
    echo "❌ User cannot write to log directory"
    echo "Error: $?"
fi
echo

echo "=== Package Installation Check ==="
if command -v rpm >/dev/null 2>&1; then
    echo "RPM packages installed:"
    rpm -qa | grep oci-dr-hpc || echo "No oci-dr-hpc packages found"
    echo
    echo "Setup script location:"
    if [ -f /usr/share/oci-dr-hpc/setup-logging.sh ]; then
        echo "✅ Setup script found at /usr/share/oci-dr-hpc/setup-logging.sh"
        ls -la /usr/share/oci-dr-hpc/setup-logging.sh
    else
        echo "❌ Setup script not found"
    fi
fi
echo

echo "=== Config File Check ==="
for config in /etc/oci-dr-hpc.yaml ~/.oci-dr-hpc.yaml; do
    if [ -f "$config" ]; then
        echo "Config found: $config"
        grep -A3 "logging:" "$config" 2>/dev/null || echo "No logging section found"
    fi
done
echo

echo "=== Recommended Actions ==="
if [ -d /var/log/oci-dr-hpc/oci-dr-hpc ]; then
    echo "1. Remove incorrect directory:"
    echo "   sudo rm -rf /var/log/oci-dr-hpc/oci-dr-hpc"
    echo
fi

echo "2. Re-run setup script:"
echo "   sudo /usr/share/oci-dr-hpc/setup-logging.sh"
echo

echo "3. Check user groups:"
echo "   groups \$(whoami)"
echo "   # User should be in wheel group on Oracle Linux"
echo