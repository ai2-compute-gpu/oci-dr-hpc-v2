# Makefile for oci-dr-hpc-v2 using FPM
APP_NAME = oci-dr-hpc-v2
VERSION ?= 1.0.0
RELEASE ?= 1

# Detect system arch and normalize to Go/FPM values
UNAME_ARCH := $(shell uname -m)
ifeq ($(UNAME_ARCH),x86_64)
  ARCH ?= amd64
else ifeq ($(UNAME_ARCH),aarch64)
  ARCH ?= arm64
else
  $(error Unsupported architecture: $(UNAME_ARCH))
endif

BUILD_DIR = build
DIST_DIR = dist

.PHONY: all all-cross clean build build-all build-amd64 build-arm64 rpm rpm-all rpm-amd64 rpm-arm64 deb deb-ubuntu deb-debian deb-all deb-ubuntu-all deb-debian-all deps install-fpm test coverage install install-dev uninstall help

all: clean test build rpm deb-ubuntu deb-debian

# Build everything for both architectures (cross-compilation)
all-cross: clean test build-all rpm-all deb-all
	@echo "Cross-compilation complete! Built packages for amd64 and arm64:"
	@echo "Binaries:"
	@ls -la $(BUILD_DIR)/$(APP_NAME)-*
	@echo "Packages:"
	@ls -la $(DIST_DIR)/*

# Build for detected architecture (default)
build:
	@echo "Building $(APP_NAME) v$(VERSION) for $(ARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -ldflags "-X main.version=$(VERSION) -s -w" -o $(BUILD_DIR)/$(APP_NAME) .

# Cross-compilation targets
build-amd64:
	@echo "Building $(APP_NAME) v$(VERSION) for amd64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION) -s -w" -o $(BUILD_DIR)/$(APP_NAME)-amd64 .

build-arm64:
	@echo "Building $(APP_NAME) v$(VERSION) for arm64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION) -s -w" -o $(BUILD_DIR)/$(APP_NAME)-arm64 .

build-all: build-amd64 build-arm64
	@echo "Built binaries for both architectures:"
	@echo "  amd64: $(BUILD_DIR)/$(APP_NAME)-amd64"
	@echo "  arm64: $(BUILD_DIR)/$(APP_NAME)-arm64"

install-fpm:
	@which fpm >/dev/null 2>&1 || (echo "Installing FPM..." && \
		if command -v dnf >/dev/null 2>&1; then \
			sudo dnf install -y ruby ruby-devel rubygems gcc make && sudo gem install fpm; \
		elif command -v yum >/dev/null 2>&1; then \
			sudo yum install -y ruby ruby-devel rubygems gcc make && sudo gem install fpm; \
		elif command -v apt-get >/dev/null 2>&1; then \
			sudo apt-get update && sudo apt-get install -y ruby ruby-dev rubygems build-essential && sudo gem install fpm; \
		else \
			echo "Unsupported package manager. Please install ruby, ruby-devel, rubygems, and build tools manually, then run 'gem install fpm'"; \
			exit 1; \
		fi)

# RPM packages for detected architecture (default)
rpm: build install-fpm
	@echo "Building RPM package with FPM (arch=$(ARCH))..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t rpm \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE) \
		--architecture $(ARCH) \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends glibc \
		--rpm-user opc \
		--rpm-group opc \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

# Cross-compilation RPM targets
rpm-amd64: build-amd64 install-fpm
	@echo "Building RPM package for amd64 with FPM..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t rpm \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE) \
		--architecture amd64 \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends glibc \
		--rpm-user opc \
		--rpm-group opc \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)-amd64=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

rpm-arm64: build-arm64 install-fpm
	@echo "Building RPM package for arm64 with FPM..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t rpm \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE) \
		--architecture arm64 \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends glibc \
		--rpm-user opc \
		--rpm-group opc \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)-arm64=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

rpm-all: rpm-amd64 rpm-arm64
	@echo "Built RPM packages for both architectures:"
	@ls -la $(DIST_DIR)/*.rpm

deb: deb-ubuntu

# DEB packages for detected architecture (default)
deb-ubuntu: build install-fpm
	@echo "Building DEB package for Ubuntu with FPM (arch=$(ARCH))..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t deb \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE)ubuntu \
		--architecture $(ARCH) \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments (Ubuntu)" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends libc6 \
		--deb-no-default-config-files \
		--deb-user ubuntu \
		--deb-group ubuntu \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

deb-debian: build install-fpm
	@echo "Building DEB package for Debian with FPM (arch=$(ARCH))..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t deb \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE)debian \
		--architecture $(ARCH) \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments (Debian)" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends libc6 \
		--deb-no-default-config-files \
		--deb-user debian \
		--deb-group debian \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

# Cross-compilation DEB Ubuntu targets
deb-ubuntu-amd64: build-amd64 install-fpm
	@echo "Building DEB package for Ubuntu amd64 with FPM..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t deb \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE)ubuntu \
		--architecture amd64 \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments (Ubuntu)" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends libc6 \
		--deb-no-default-config-files \
		--deb-user ubuntu \
		--deb-group ubuntu \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)-amd64=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

deb-ubuntu-arm64: build-arm64 install-fpm
	@echo "Building DEB package for Ubuntu arm64 with FPM..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t deb \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE)ubuntu \
		--architecture arm64 \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments (Ubuntu)" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends libc6 \
		--deb-no-default-config-files \
		--deb-user ubuntu \
		--deb-group ubuntu \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)-arm64=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

deb-ubuntu-all: deb-ubuntu-amd64 deb-ubuntu-arm64
	@echo "Built Ubuntu DEB packages for both architectures:"
	@ls -la $(DIST_DIR)/*ubuntu*.deb

# Cross-compilation DEB Debian targets  
deb-debian-amd64: build-amd64 install-fpm
	@echo "Building DEB package for Debian amd64 with FPM..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t deb \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE)debian \
		--architecture amd64 \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments (Debian)" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends libc6 \
		--deb-no-default-config-files \
		--deb-user debian \
		--deb-group debian \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)-amd64=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

deb-debian-arm64: build-arm64 install-fpm
	@echo "Building DEB package for Debian arm64 with FPM..."
	@mkdir -p $(DIST_DIR)
	fpm -s dir -t deb \
		--name $(APP_NAME) \
		--version $(VERSION) \
		--iteration $(RELEASE)debian \
		--architecture arm64 \
		--description "Oracle Cloud Infrastructure - OCI GPU Diagnostic and Repair tool for HPC environments (Debian)" \
		--url "https://www.oracle.com/ai-infrastructure/" \
		--license "Oracle" \
		--maintainer "Bob R Booth <bob.r.booth@oracle.com>" \
		--depends libc6 \
		--deb-no-default-config-files \
		--deb-user debian \
		--deb-group debian \
		--after-install scripts/setup-logging.sh \
		--package $(DIST_DIR) \
		$(BUILD_DIR)/$(APP_NAME)-arm64=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		internal/shapes/shapes.json=/etc/oci-dr-hpc-shapes.json \
		configs/recommendations.json=/usr/share/oci-dr-hpc/recommendations.json \
		internal/test_limits/test_limits.json=/etc/oci-dr-hpc-test-limits.json \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh

deb-debian-all: deb-debian-amd64 deb-debian-arm64
	@echo "Built Debian DEB packages for both architectures:"
	@ls -la $(DIST_DIR)/*debian*.deb

deb-all: deb-ubuntu-all deb-debian-all
	@echo "Built all DEB packages for both architectures:"
	@ls -la $(DIST_DIR)/*.deb

test:
	@echo "Running unit tests..."
	go test -v ./...

coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	go test -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"
	go tool cover -func=$(BUILD_DIR)/coverage.out

install: build
	@echo "Installing $(APP_NAME) system-wide..."
	@sudo mkdir -p /usr/bin
	@sudo mkdir -p /usr/share/oci-dr-hpc
	@sudo mkdir -p /etc/oci-dr-hpc
	@sudo install -m 755 $(BUILD_DIR)/$(APP_NAME) /usr/bin/
	@sudo install -m 644 configs/recommendations.json /usr/share/oci-dr-hpc/
	@sudo install -m 644 internal/test_limits/test_limits.json /etc/oci-dr-hpc-test-limits.json
	@if [ ! -f /etc/oci-dr-hpc/recommendations.json ]; then \
		sudo install -m 644 configs/recommendations.json /etc/oci-dr-hpc/; \
	fi
	@echo "Installation complete!"
	@echo "Binary: /usr/bin/$(APP_NAME)"
	@echo "Default config: /usr/share/oci-dr-hpc/recommendations.json"
	@echo "System config: /etc/oci-dr-hpc/recommendations.json"
	@echo "Test limits config: /etc/oci-dr-hpc-test-limits.json"

install-dev: build
	@echo "Installing $(APP_NAME) for development..."
	@mkdir -p ~/.local/bin
	@mkdir -p ~/.config/oci-dr-hpc
	@cp $(BUILD_DIR)/$(APP_NAME) ~/.local/bin/
	@cp configs/recommendations.json ~/.config/oci-dr-hpc/
	@cp internal/test_limits/test_limits.json ~/.config/oci-dr-hpc/
	@echo "Development installation complete!"
	@echo "Binary: ~/.local/bin/$(APP_NAME)"
	@echo "Config: ~/.config/oci-dr-hpc/recommendations.json"
	@echo "Test limits config: ~/.config/oci-dr-hpc/test_limits.json"
	@echo "Make sure ~/.local/bin is in your PATH"

uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	@sudo rm -f /usr/bin/$(APP_NAME)
	@sudo rm -f /usr/share/oci-dr-hpc/recommendations.json
	@echo "Note: Custom configs in /etc/oci-dr-hpc/ were preserved"
	@echo "Remove manually if needed: sudo rm -rf /etc/oci-dr-hpc/"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)

help:
	@echo "OCI DR HPC v2 Build System"
	@echo "=========================="
	@echo ""
	@echo "🏗️  BUILD TARGETS:"
	@echo "  build          - Build for detected architecture ($(ARCH))"
	@echo "  build-amd64    - Cross-compile for amd64 (x86_64)"
	@echo "  build-arm64    - Cross-compile for arm64 (aarch64)"
	@echo "  build-all      - Build for both amd64 and arm64"
	@echo ""
	@echo "📦 PACKAGE TARGETS:"
	@echo "  rpm            - Build RPM for detected architecture"
	@echo "  rpm-amd64      - Build RPM for amd64"
	@echo "  rpm-arm64      - Build RPM for arm64"
	@echo "  rpm-all        - Build RPM for both architectures"
	@echo ""
	@echo "  deb-ubuntu     - Build Ubuntu DEB for detected architecture"
	@echo "  deb-ubuntu-amd64   - Build Ubuntu DEB for amd64"
	@echo "  deb-ubuntu-arm64   - Build Ubuntu DEB for arm64"
	@echo "  deb-ubuntu-all     - Build Ubuntu DEB for both architectures"
	@echo ""
	@echo "  deb-debian     - Build Debian DEB for detected architecture"
	@echo "  deb-debian-amd64   - Build Debian DEB for amd64"
	@echo "  deb-debian-arm64   - Build Debian DEB for arm64"
	@echo "  deb-debian-all     - Build Debian DEB for both architectures"
	@echo ""
	@echo "  deb-all        - Build all DEB packages for both architectures"
	@echo ""
	@echo "🚀 COMBINED TARGETS:"
	@echo "  all            - Build and package for detected architecture"
	@echo "  all-cross      - Build and package for both amd64 and arm64"
	@echo ""
	@echo "🧪 TESTING:"
	@echo "  test           - Run unit tests"
	@echo "  coverage       - Run tests with coverage report"
	@echo ""
	@echo "⚙️  INSTALLATION:"
	@echo "  install        - Install system-wide (requires sudo)"
	@echo "  install-dev    - Install for development (user-local)"
	@echo "  uninstall      - Remove system installation"
	@echo ""
	@echo "🛠️  UTILITIES:"
	@echo "  clean          - Remove build artifacts"
	@echo "  install-fpm    - Install FPM package builder"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "🏛️  ARCHITECTURE INFO:"
	@echo "  Detected: $(ARCH) ($(UNAME_ARCH))"
	@echo "  Supported: amd64 (x86_64), arm64 (aarch64)"