#!/bin/bash

# OCI IMDS Test Runner Script
# This script builds and runs the IMDS test program

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "=== OCI IMDS Test Runner ==="
echo "Project root: ${PROJECT_ROOT}"
echo

# Check if we're in the right directory
if [ ! -f "${PROJECT_ROOT}/go.mod" ]; then
    echo "❌ Error: Cannot find go.mod in project root"
    echo "   Please run this script from the oci-dr-hpc-v2 project directory"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "   Please install Go 1.21.5 or higher"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: ${GO_VERSION}"

# Change to project root
cd "${PROJECT_ROOT}"

# Build the test program
echo "🔨 Building IMDS test program..."
go build -o scratch/test_imds_standalone scratch/test_imds_standalone.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

echo
echo "🚀 Running IMDS test..."
echo "   Note: This must be run on an OCI compute instance"
echo "   If not on OCI, the test will exit with an error"
echo

# Run the test program
./scratch/test_imds_standalone

echo
echo "=== Test Complete ==="
echo "📝 Log files (if configured) are in /var/log/oci-dr-hpc/"
echo "🧹 Cleaning up build artifacts..."
rm -f scratch/test_imds_standalone

echo "✅ Done!" 