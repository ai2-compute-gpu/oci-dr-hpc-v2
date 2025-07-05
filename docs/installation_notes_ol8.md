```markdown
# Installing Ruby 3.2.2 and FPM on Oracle Linux 8.10

## Background

This setup is required to resolve a build failure when running `make` in a project that depends on the `fpm` gem.

During the `make install-fpm` step, the build fails with the following error:

```

dotenv requires Ruby version >= 3.0. The current ruby version is 2.5.0.

````

Oracle Linux 8.10 ships with an outdated system Ruby (2.5), which is incompatible with modern gems like `dotenv`, a dependency of `fpm`. This guide provides a clean way to install Ruby 3.2.2 via `rbenv`, along with `fpm`, without affecting the system Ruby.

---

## Step 1: Install Build Dependencies

Run the following to install required system packages:

```bash
sudo dnf install -y \
  gcc gcc-c++ make patch bzip2 \
  openssl-devel libffi-devel readline-devel zlib-devel \
  wget tar git \
  which \
  dnf-plugins-core
````

## Step 2: Enable CodeReady Builder Repository

Enable the repository that contains development headers:

```bash
sudo dnf config-manager --set-enabled ol8_codeready_builder
sudo dnf clean all && sudo dnf makecache
sudo dnf install -y libyaml-devel
```

## Step 3: Install rbenv and ruby-build

Install `rbenv` and its plugin `ruby-build` to manage Ruby versions:

```bash
# Clone rbenv
git clone https://github.com/rbenv/rbenv.git ~/.rbenv
echo 'export PATH="$HOME/.rbenv/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(rbenv init - bash)"' >> ~/.bashrc
source ~/.bashrc

# Clone ruby-build
git clone https://github.com/rbenv/ruby-build.git ~/.rbenv/plugins/ruby-build
```

## Step 4: Install Ruby 3.2.2 with rbenv

Use the following command to build and install Ruby 3.2.2, disabling DTrace to avoid build errors:

```bash
RUBY_CONFIGURE_OPTS="--disable-dtrace" rbenv install 3.2.2
rbenv global 3.2.2
```

Verify the installation:

```bash
ruby -v
# Expected output: ruby 3.2.2
```

## Step 5: Install FPM

Install `fpm` via RubyGems:

```bash
gem install fpm
```

Verify the installation:

```bash
fpm --version
```

## Optional: Additional Tools for Packaging

To build RPM packages with `fpm`, install:

```bash
sudo dnf install -y rpm-build
```

If building Debian packages on a Debian-based system:

```bash
sudo apt install -y dpkg-dev
```

## Notes

* Ruby is installed in `~/.rbenv/versions/3.2.2`
* This method avoids interfering with the system Ruby
* Suitable for development environments, CI/CD pipelines, or packaging workflows where `fpm` is used
