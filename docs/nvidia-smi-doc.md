# NVIDIA System Management Interface (`nvidia-smi`)

## NAME

`nvidia-smi` – NVIDIA System Management Interface program

ref: https://developer.download.nvidia.com/compute/DCGM/docs/nvidia-smi-367.38.pdf

## SYNOPSIS

```bash
nvidia-smi [OPTION1 [ARG1]] [OPTION2 [ARG2]] ...
```

## DESCRIPTION

The `nvidia-smi` (NVSMI) utility provides monitoring and management capabilities for NVIDIA GPUs (Tesla, Quadro, GRID, GeForce – Fermi architecture and newer). It supports Linux and 64-bit Windows. Output formats include human-readable, CSV, and XML.

> **Note:** Prefer using NVML (NVIDIA Management Library) or Python bindings for tools needing backward compatibility.

* NVML: [NVIDIA NVML SDK](http://developer.nvidia.com/nvidia-management-library-nvml/)
* Python bindings: [nvidia-ml-py](http://pypi.python.org/pypi/nvidia-ml-py/)

---

## GENERAL OPTIONS

| Option       | Description                      |
| ------------ | -------------------------------- |
| `-h, --help` | Print usage information and exit |

---

## SUMMARY OPTIONS

| Option | Description               |
| ------ | ------------------------- |
| `-L`   | List GPUs and their UUIDs |

---

## QUERY OPTIONS

| Option         | Description                                          |
| -------------- | ---------------------------------------------------- |
| `-q`           | Display full GPU/Unit information                    |
| `-u`           | Display Unit data instead of GPU data                |
| `-i=ID`        | Specify device by index, UUID, or PCI bus ID         |
| `-f FILE`      | Redirect output to a file                            |
| `-x`           | Output in XML format                                 |
| `--dtd`        | Embed DTD in XML                                     |
| `--debug=FILE` | Create encrypted debug log                           |
| `-d TYPE`      | Show selected data types (e.g., MEMORY, UTILIZATION) |
| `-l SEC`       | Repeat query every N seconds                         |

---

## SELECTIVE QUERY

```bash
--query-gpu=pci.bus_id,persistence_mode --format=csv,noheader
```

* Other query options:

    * `--query-supported-clocks`
    * `--query-compute-apps`
    * `--query-accounted-apps`
    * `--query-retired-pages`

---

## DEVICE MODIFICATION OPTIONS

| Option             | Description                                        |
| ------------------ | -------------------------------------------------- |
| `-pm MODE`         | Set persistence mode (Linux only)                  |
| `-e CONFIG`        | Set ECC mode (requires reboot)                     |
| `-p TYPE`          | Reset ECC error counters (VOLATILE or AGGREGATE)   |
| `-c MODE`          | Set compute mode                                   |
| `--gom=MODE`       | Set GPU operation mode (ALL\_ON, COMPUTE, LOW\_DP) |
| `-r`               | Reset GPU (requires no processes using GPU)        |
| `-ac MEM,GRAPHICS` | Set application clocks                             |
| `-rac`             | Reset application clocks to default                |
| `-acp MODE`        | Set permission for clock changes                   |
| `-pl WATTS`        | Set power limit (Kepler+)                          |
| `-am MODE`         | Enable/Disable GPU accounting                      |
| `-caa`             | Clear accounted apps                               |
| `--auto-boost-*`   | Control auto-boost features                        |

---

## ADDITIONAL COMMANDS

* `stats`: Display GPU stats
* `topo`: Show topology between GPUs
* `nvlink`: Show NVLink info
* `clocks`: Control clocking
* `vgpu`: Show vGPU info
* `dmon`: Device monitor tool (interactive)
* `daemon`: Background monitor with logs
* `replay`: Replay collected stats from daemon
* `pmon`: Process-level monitoring
* `--help` on any subcommand provides detailed usage

---

## EXAMPLES

```bash
nvidia-smi -q
nvidia-smi --format=csv,noheader --query-gpu=uuid,persistence_mode
nvidia-smi -q -d ECC,POWER -i 0 -l 10 -f out.log
nvidia-smi -i 0 --applications-clocks 2500,745
nvidia-smi -q -u -x --dtd
nvidia-smi daemon -i 0,1 -s pucvmet -d 5 -p /tmp
```

---

## RETURN CODES

| Code | Meaning                                 |
| ---- | --------------------------------------- |
| 0    | Success                                 |
| 2    | Invalid argument                        |
| 3    | Unsupported operation on device         |
| 4    | Insufficient permissions                |
| 9    | NVIDIA driver not loaded                |
| 12   | NVML shared library load failure        |
| 15   | GPU inaccessible ("fallen off the bus") |

---

## GPU ATTRIBUTES (sample list)

* Driver Version
* UUID
* PCI Bus ID
* Power Limit, Draw, Management
* Clocks (Graphics, Memory, SM)
* Temperature
* Fan Speed
* Utilization
* ECC Errors
* Page Retirement
* Performance State

---

## NOTES

* On Linux, `nvidia-smi` may modify device files if run as root.
* Deprecated flags: `-a` (use `-q`), `-g` (use `-i`)

---

Let me know if you’d like this saved as a `.md` file or want a shorter version.
