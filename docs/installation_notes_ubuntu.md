---

````markdown
# 🛠️ Building `oci-dr-hpc-v2` on Ubuntu

This guide outlines the steps required to build the `oci-dr-hpc-v2` CLI tool on a clean Ubuntu system.

---

## 1. Install Go (v1.22.3)

Remove older Go (optional):

```bash
sudo apt remove -y golang-go
````

Download and install Go based on your system architecture:

### For x86_64 (amd64) systems:

```bash
cd /tmp
wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
```

### For ARM64 (aarch64) systems:

```bash
cd /tmp
wget https://go.dev/dl/go1.22.3.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.3.linux-arm64.tar.gz
```

### Auto-detect architecture (recommended):

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

wget https://go.dev/dl/go1.22.3.linux-${GO_ARCH}.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.3.linux-${GO_ARCH}.tar.gz
```

Update your `PATH`:

```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

Verify installation:

```bash
go version
# Expected output for x86_64: go version go1.22.3 linux/amd64
# Expected output for ARM64: go version go1.22.3 linux/arm64
```

---

## 2. Clone the Repository

Use your GitHub SSH key to clone the project:

```bash
git clone git@github.com:ai2-compute-gpu/oci-dr-hpc-v2.git
cd oci-dr-hpc-v2
```

---

## 3. Install `fpm` for Packaging (Optional)

`fpm` is required if you want to package the CLI as a `.deb` or `.rpm`.

Install Ruby and build dependencies:

```bash
sudo apt update
sudo apt install -y ruby ruby-dev build-essential
```

Install `fpm`:

```bash
sudo gem install --no-document fpm
```

---

## 4. Build the Project

Run the `make` command from the project root:

```bash
make
```

This will:

* Run unit tests
* Build the binary at `build/oci-dr-hpc-v2`
* Optionally generate packages via `fpm`

---

## 5. Run the CLI

After a successful build, run the CLI:

```bash
./build/oci-dr-hpc-v2 --help
```

---

## Notes

* If you only need the binary and want to skip packaging, you can comment or remove the `install-fpm` target in the `Makefile`.
* Ensure your GitHub SSH key is loaded with `ssh-add ~/.ssh/github` if using private repositories.

```