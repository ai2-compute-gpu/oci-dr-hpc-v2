# Makefile for oci-dr-hpc-v2 using FPM
APP_NAME = oci-dr-hpc-v2
VERSION ?= 1.0.0
RELEASE ?= 1
ARCH ?= x86_64

BUILD_DIR = build
DIST_DIR = dist

.PHONY: all clean build rpm deb deb-ubuntu deb-debian deps install-fpm test coverage

all: clean test build rpm deb-ubuntu deb-debian

build:
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)/var/log/oci-dr-hpc
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION) -s -w" -o $(BUILD_DIR)/$(APP_NAME) .

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

rpm: build install-fpm
	@echo "Building RPM package with FPM..."
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
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh \
		$(BUILD_DIR)/var/log/oci-dr-hpc=/var/log/oci-dr-hpc

deb: deb-ubuntu

deb-ubuntu: build install-fpm
	@echo "Building DEB package for Ubuntu with FPM..."
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
		$(BUILD_DIR)/$(APP_NAME)=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh \
		$(BUILD_DIR)/var/log/oci-dr-hpc=/var/log/oci-dr-hpc

deb-debian: build install-fpm
	@echo "Building DEB package for Debian with FPM..."
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
		$(BUILD_DIR)/$(APP_NAME)=/usr/bin/$(APP_NAME) \
		config/oci-dr-hpc.yaml=/etc/oci-dr-hpc.yaml \
		scripts/setup-logging.sh=/usr/share/oci-dr-hpc/setup-logging.sh \
		$(BUILD_DIR)/var/log/oci-dr-hpc=/var/log/oci-dr-hpc

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

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)