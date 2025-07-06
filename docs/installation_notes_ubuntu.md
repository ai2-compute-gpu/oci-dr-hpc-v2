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

Download and install Go:

```bash
cd /tmp
wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
```

Update your `PATH`:

```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

Verify installation:

```bash
go version
# Expected output: go version go1.22.3 linux/amd64
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