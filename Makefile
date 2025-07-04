# Makefile for oci-dr-hpc-v2
# Builds RPM and DEB packages for Linux

# Variables
APP_NAME = oci-dr-hpc-v2
VERSION ?= 1.0.0
RELEASE ?= 1
ARCH ?= x86_64
GO_VERSION = 1.21.5

# Build directories
BUILD_DIR = build
DIST_DIR = dist
RPM_BUILD_DIR = $(BUILD_DIR)/rpmbuild
DEB_BUILD_DIR = $(BUILD_DIR)/deb

# Go build flags
GO_LDFLAGS = -ldflags "-X main.version=$(VERSION) -s -w"
GO_BUILD_FLAGS = -a -installsuffix cgo

.PHONY: all clean build rpm deb deps

all: clean build rpm deb

# Build the Go binary
build:
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

# Install build dependencies
deps:
	@echo "Installing build dependencies..."
	@which rpm-build >/dev/null 2>&1 || (echo "Installing rpm-build..." && sudo yum install -y rpm-build rpmdevtools || sudo apt-get install -y rpm)
	@which dpkg-deb >/dev/null 2>&1 || (echo "Installing dpkg-deb..." && sudo apt-get install -y dpkg-dev)

# Build RPM package
rpm: build
	@echo "Building RPM package..."
	@mkdir -p $(RPM_BUILD_DIR)/BUILD
	@mkdir -p $(RPM_BUILD_DIR)/BUILDROOT
	@mkdir -p $(RPM_BUILD_DIR)/RPMS
	@mkdir -p $(RPM_BUILD_DIR)/SOURCES
	@mkdir -p $(RPM_BUILD_DIR)/SPECS
	@mkdir -p $(RPM_BUILD_DIR)/SRPMS
	@mkdir -p $(RPM_BUILD_DIR)/BUILDROOT/$(APP_NAME)-$(VERSION)-$(RELEASE).$(ARCH)/usr/bin
	@mkdir -p $(RPM_BUILD_DIR)/BUILDROOT/$(APP_NAME)-$(VERSION)-$(RELEASE).$(ARCH)/etc/$(APP_NAME)
	@mkdir -p $(RPM_BUILD_DIR)/BUILDROOT/$(APP_NAME)-$(VERSION)-$(RELEASE).$(ARCH)/usr/share/doc/$(APP_NAME)
	@cp $(BUILD_DIR)/$(APP_NAME) $(RPM_BUILD_DIR)/BUILDROOT/$(APP_NAME)-$(VERSION)-$(RELEASE).$(ARCH)/usr/bin/
	@cp README.md $(RPM_BUILD_DIR)/BUILDROOT/$(APP_NAME)-$(VERSION)-$(RELEASE).$(ARCH)/usr/share/doc/$(APP_NAME)/ 2>/dev/null || true
	@echo "Name: $(APP_NAME)" > $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "Version: $(VERSION)" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "Release: $(RELEASE)" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "Summary: Oracle Cloud Infrastructure Diagnostic and Repair tool for HPC environments" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "License: Oracle" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "Group: Applications/System" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "BuildArch: $(ARCH)" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "Requires: glibc" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "%description" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "Oracle Cloud Infrastructure Diagnostic and Repair tool for HPC environments with GPU and RDMA support." >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "%files" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "/usr/bin/$(APP_NAME)" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "/usr/share/doc/$(APP_NAME)/*" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "%changelog" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "* $(shell date '+%a %b %d %Y') Builder <builder@oracle.com> - $(VERSION)-$(RELEASE)" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@echo "- Initial package build" >> $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@rpmbuild --define "_topdir $(PWD)/$(RPM_BUILD_DIR)" -bb $(RPM_BUILD_DIR)/SPECS/$(APP_NAME).spec
	@mkdir -p $(DIST_DIR)
	@cp $(RPM_BUILD_DIR)/RPMS/$(ARCH)/$(APP_NAME)-$(VERSION)-$(RELEASE).$(ARCH).rpm $(DIST_DIR)/
	@echo "RPM package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$(RELEASE).$(ARCH).rpm"

# Build DEB package
deb: build
	@echo "Building DEB package..."
	@mkdir -p $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN
	@mkdir -p $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/bin
	@mkdir -p $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/etc/$(APP_NAME)
	@mkdir -p $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/share/doc/$(APP_NAME)
	@cp $(BUILD_DIR)/$(APP_NAME) $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/bin/
	@cp README.md $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/share/doc/$(APP_NAME)/ 2>/dev/null || true
	@echo "Package: $(APP_NAME)" > $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@echo "Version: $(VERSION)-$(RELEASE)" >> $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@echo "Section: utils" >> $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@echo "Priority: optional" >> $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@echo "Architecture: amd64" >> $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@echo "Depends: libc6" >> $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@echo "Maintainer: Oracle <support@oracle.com>" >> $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@echo "Description: Oracle Cloud Infrastructure Diagnostic and Repair tool for HPC environments" >> $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@echo " Oracle Cloud Infrastructure Diagnostic and Repair tool for HPC environments with GPU and RDMA support." >> $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	@chmod 755 $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/bin/$(APP_NAME)
	@dpkg-deb --build $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64
	@mkdir -p $(DIST_DIR)
	@cp $(DEB_BUILD_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64.deb $(DIST_DIR)/
	@echo "DEB package created: $(DIST_DIR)/$(APP_NAME)_$(VERSION)-$(RELEASE)_amd64.deb"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)

# Show help
help:
	@echo "Available targets:"
	@echo "  all     - Build binary and create both RPM and DEB packages"
	@echo "  build   - Build the Go binary only"
	@echo "  rpm     - Create RPM package"
	@echo "  deb     - Create DEB package"
	@echo "  deps    - Install build dependencies"
	@echo "  clean   - Clean build artifacts"
	@echo "  help    - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION - Package version (default: 1.0.0)"
	@echo "  RELEASE - Package release (default: 1)"
	@echo "  ARCH    - Target architecture (default: x86_64)"