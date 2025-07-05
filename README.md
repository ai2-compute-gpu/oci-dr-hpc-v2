# oci-dr-hpc-v2

@brbooth
@rekharoy
@bhrajan

## Bash Scripts

The Bash and Python scripts in the `/scripts` are intended to be equivalents of the Dr HPC GoLang CLI. 

## GoLang

### Usage


### How to Build

If you are cloning this repository and building from source, ensure you have the following dependencies installed for the respective platform.

#### Oracle Linux


##### Prerequisites

| Operating System | Compatible |
| -------- | ------- |
| `Oracle Linux 9.5`    |  :white_check_mark:   |
| `Ubuntu 22.04`   | :white_check_mark:   |


Build dependencies
| Dependency | Version |
| -------- | ------- |
| `git`    |  v2.47.1   |
| `make`   | v4.3    |
| `golang` | v1.21.5     |


Dr HPC uses the below to run tests and healthchecks
Runtime dependencies
| Dependency | Version | Additional Notes |
| -------- | ------- | ----- |
| `nvidia-smi`|      |       |
| `lspci`   |        |       |
| `mlxconfig` |      |       |


### Build Dr HPC CLI From Source

1) Assuming the host has access to publicly available YUM/RPM repositories, these can be installed by running:
```bash
$ dnf update && dnf install -y git build-essential
```

2) Visit https://go.dev/dl/ to download and install GoLang for your operating system.

An example GoLang installation can be followed below.

```bash
$ dnf install wget
$ wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
$ tar -xvf go1.21.5.linux-amd64.tar.gz
$ mv go /usr/local
$ export PATH=$PATH:/usr/local/go/bin
```

3) Build Dr HPC CLI
```bash
$ git clone https://github.com/ai2-compute-gpu/oci-dr-hpc-v2.git
$ cd oci-dr-hpc-v2
$ make build 
```

The CLI executable will be in the `/build` directory.
