# 🛠️ Building `oci-dr-hpc-v2` on Ubuntu

This guide outlines the steps required to build the `oci-dr-hpc-v2` CLI tool on Ubuntu systems with support for both AMD64 and ARM64 architectures.

---

## 📋 Prerequisites

- Ubuntu 18.04+ (tested on Ubuntu 20.04, 22.04, 24.04)
- Internet connection for downloading dependencies
- Git for repository access
- Sudo privileges for package installation

---

## 1. Install Go (v1.21.5+)

### Remove Older Go Installation (if needed)

```bash
sudo apt remove -y golang-go
sudo rm -rf /usr/local/go
```

### Download and Install Go

**Auto-detect architecture (recommended):**

```bash
cd /tmp
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    GO_ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    GO_ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Download latest stable Go (update version as needed)
GO_VERSION="1.22.3"
wget https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-${GO_ARCH}.tar.gz
```

**Manual installation by architecture:**

```bash
# For x86_64 (Intel/AMD) systems:
cd /tmp
wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz

# For ARM64 (AWS Graviton, etc.) systems:
cd /tmp
wget https://go.dev/dl/go1.22.3.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.22.3.linux-arm64.tar.gz
```

### Configure Go Environment

```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc
```

### Verify Installation

```bash
go version
# Expected output for x86_64: go version go1.22.3 linux/amd64
# Expected output for ARM64: go version go1.22.3 linux/arm64

# Check your system architecture
uname -m
# x86_64 = Intel/AMD 64-bit
# aarch64 = ARM 64-bit
```

---

## 2. Install Build Dependencies

### Install Required System Packages

```bash
sudo apt update
sudo apt install -y \
    git \
    build-essential \
    curl \
    wget
```

### Install FPM for Package Building (Optional)

FPM is required only if you want to create `.deb` or `.rpm` packages:

```bash
# Install Ruby and development tools
sudo apt install -y ruby ruby-dev rubygems

# Install FPM
sudo gem install --no-document fpm
```

### Install RPM Tools (Optional)

If you need to build RPM packages on Ubuntu:

```bash
sudo apt install -y rpm
```

---

## 3. Clone the Repository

```bash
# Clone the repository (adjust URL as needed)
git clone <repository-url> oci-dr-hpc-v2
cd oci-dr-hpc-v2

# Verify you're in the correct directory
ls -la
# Should see: Makefile, main.go, cmd/, internal/, etc.
```

---

## 4. Build the Project

### Quick Build (Detected Architecture)

```bash
# Build for your current architecture
make build

# Run unit tests
make test

# Build and test everything
make
```

### Cross-Platform Building

The build system supports cross-compilation for multiple architectures:

```bash
# Build for specific architectures
make build-amd64    # Build for x86_64 systems
make build-arm64    # Build for ARM64 systems
make build-all      # Build for both architectures

# View all available build targets
make help
```

### Package Building

```bash
# Build packages for detected architecture
make rpm            # RPM package
make deb-ubuntu     # Ubuntu DEB package

# Cross-platform package building
make rpm-all        # RPM for both architectures
make deb-all        # DEB for both architectures
make all-cross      # Everything for both architectures
```

---

## 5. Installation Options

### Development Installation

Install for current user only:

```bash
make install-dev
# Binary: ~/.local/bin/oci-dr-hpc-v2
# Config: ~/.config/oci-dr-hpc/recommendations.json
```

### System-Wide Installation

Install for all users:

```bash
sudo make install
# Binary: /usr/bin/oci-dr-hpc-v2
# Config: /usr/share/oci-dr-hpc/recommendations.json
```

### Package Installation

Install from built packages:

```bash
# Install DEB package
sudo dpkg -i dist/oci-dr-hpc-v2-*.deb

# Install RPM package (if rpm tools installed)
sudo rpm -i dist/oci-dr-hpc-v2-*.rpm
```

---

## 6. Verify Installation

### Test the Binary

```bash
# If using development installation
~/.local/bin/oci-dr-hpc-v2 --version

# If using system installation
oci-dr-hpc-v2 --version

# Test help output
oci-dr-hpc-v2 --help
```

### Verify Cross-Compilation (Optional)

```bash
# Check built binaries
ls -la build/
# Should see: oci-dr-hpc-v2-amd64, oci-dr-hpc-v2-arm64 (if cross-compiled)

# Verify architectures
file build/oci-dr-hpc-v2-amd64
file build/oci-dr-hpc-v2-arm64
```

---

## 🔧 Troubleshooting

### Common Issues

**Go not found after installation:**
```bash
# Ensure Go is in PATH
which go
# If not found, reload shell configuration:
source ~/.bashrc
```

**Permission denied during gem install:**
```bash
# Use user-local gem installation
gem install --user-install fpm
export PATH="$PATH:$(ruby -e 'print Gem.user_dir')/bin"
```

**Build fails with module errors:**
```bash
# Clean and retry
go clean -modcache
go mod download
make clean
make build
```

**Cross-compilation fails:**
```bash
# Ensure CGO is disabled for static builds
export CGO_ENABLED=0
make build-all
```

### Build System Help

```bash
# View all available targets
make help

# Clean build artifacts
make clean

# Run only tests
make test

# Build with coverage
make coverage
```

---

## 📁 Build Outputs

After successful compilation, you'll find:

```
build/
├── oci-dr-hpc-v2           # Native architecture binary
├── oci-dr-hpc-v2-amd64     # x86_64 binary (if cross-compiled)
└── oci-dr-hpc-v2-arm64     # ARM64 binary (if cross-compiled)

dist/                        # Packages (if built)
├── oci-dr-hpc-v2-*.deb     # DEB packages
└── oci-dr-hpc-v2-*.rpm     # RPM packages
```

---

## 🚀 Quick Start Commands

```bash
# Complete build process
sudo apt update
sudo apt install -y git build-essential
cd /tmp && wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc

# Clone and build
git clone <repository-url> oci-dr-hpc-v2
cd oci-dr-hpc-v2
make build

# Test installation
./build/oci-dr-hpc-v2 --version
```

---

## 📚 Additional Resources

- **Project README**: `../README.md` - Comprehensive project documentation
- **Makefile Help**: `make help` - View all build targets
- **Go Documentation**: https://golang.org/doc/install
- **FPM Documentation**: https://fpm.readthedocs.io/