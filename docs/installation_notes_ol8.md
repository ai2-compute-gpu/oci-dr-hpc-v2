```markdown
# Installing Go, Ruby 3.2.2, and FPM on Oracle Linux 8.10

## Background

This setup is required to resolve issues encountered when building a Go-based project on Oracle Linux 8.10:

1. `make` fails with:
```

/bin/sh: go: command not found

```
This indicates that the Go toolchain is not installed.

2. `make install-fpm` fails with:
```

dotenv requires Ruby version >= 3.0. The current ruby version is 2.5.0.

````
This occurs because the default Ruby version (2.5.0) is too old for modern Ruby gems used by the `fpm` packaging tool.

This guide installs the correct versions of Go and Ruby to allow the build system and packaging steps to work cleanly.

---

## Step 1: Install Go (1.21.x or later)

Download and install Go:

```bash
cd /tmp
curl -LO https://go.dev/dl/go1.21.10.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.10.linux-amd64.tar.gz
````

Add Go to your shell environment:

```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc
```

Verify:

```bash
go version
# Expected: go version go1.21.10 linux/amd64
```

---

## Step 2: Install Build Dependencies for Ruby

```bash
sudo dnf install -y \
  gcc gcc-c++ make patch bzip2 \
  openssl-devel libffi-devel readline-devel zlib-devel \
  wget tar git which dnf-plugins-core
```

Enable CodeReady Builder repo:

```bash
sudo dnf config-manager --set-enabled ol8_codeready_builder
sudo dnf clean all && sudo dnf makecache
sudo dnf install -y libyaml-devel
```

---

## Step 3: Install rbenv and ruby-build

```bash
git clone https://github.com/rbenv/rbenv.git ~/.rbenv
echo 'export PATH="$HOME/.rbenv/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(rbenv init - bash)"' >> ~/.bashrc
source ~/.bashrc

git clone https://github.com/rbenv/ruby-build.git ~/.rbenv/plugins/ruby-build
```

---

## Step 4: Install Ruby 3.2.2

```bash
RUBY_CONFIGURE_OPTS="--disable-dtrace" rbenv install 3.2.2
rbenv global 3.2.2
```

Verify:

```bash
ruby -v
# Expected: ruby 3.2.2
```

---

## Step 5: Install FPM

```bash
gem install fpm
```

Verify:

```bash
fpm --version
```

---

## Optional: RPM Packaging Support

If your project builds `.rpm` files using `fpm`, also install:

```bash
sudo dnf install -y rpm-build
```

---

## Notes

* Go is installed at `/usr/local/go`
* Ruby is installed via rbenv in `~/.rbenv/versions/3.2.2`
* This setup avoids modifying system-level Ruby and Go installations
* Compatible with Oracle Linux 8.10 for CI/CD or local dev use

