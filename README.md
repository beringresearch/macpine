[![Go Report Card](https://goreportcard.com/badge/github.com/beringresearch/macpine)](https://goreportcard.com/report/github.com/beringresearch/macpine)

# Lightweight Alpine VMs on MacOS


## Motivation
The goal of this project is to enable MacOS users to:

1. Easily spin up and manage lightweight Linux environments.
2. Use tiny VMs to take advantage of containerisation technologies, including [LXD](https://linuxcontainers.org/lxd/introduction/) and Docker.
3. Build and test software on x86_64 and arm64 systems.

# Installation

Ensure that QEMU is available to your system:

```bash
brew install qemu
```

## Install from source

```bash
git clone https://github.com/beringresearch/macpine
cd macpine
make all
```

# Getting Started

To launch a brand new VM:

```bash
alpine launch #launches with default parameters
alpine launch -a aarch64 #create an aarch64 instance
alpine launch -d 10G -c 4 -m 2048 #create a machine with a 10GB disk, 4 cpus and 2GB of RAM

```

Access VM via ssh:

```bash
alpine launch -s 22 #launch VM and expose SSH port to host port 22
ssh root@localhost -p 22 #password: root
```

Expose additional VM ports to host:

```bash
alpine launch -s 23 -p 8888,5432 #launch VM, exposes SSH to host port 23 and forwards VM ports 8888 and 5432 to host ports 8888 and 5432
```

VMs can be easily packaged for export and re-use as tar.gz files:

```bash
alpine list

NAME                STATUS      SSH    PORTS ARCH        PID 
forthright-hook     Running     23           aarch64     91598
hot-cow             Running     22           x86_64      82361
```

```bash
alpine publish hot-cow
```

This will create a file hot-cow.tar.gz which can be imported as:

```bash
alpine import hot-cow.tar.gz
```