[![Go Report Card](https://goreportcard.com/badge/github.com/beringresearch/macpine)](https://goreportcard.com/report/github.com/beringresearch/macpine)

# Lightweight Alpine VMs on MacOS


## Motivation
The goal of this project is to enable MacOS users to:

1. Easily spin up and manage lightweight Linux environments.
2. Use tiny VMs to take advantage of containerisation technologies, including [LXD](https://linuxcontainers.org/lxd/introduction/) and Docker.
3. Build and test software on x86_64 and arm64 systems.

# Installation

## Install from source

```bash
git clone https://github.com/beringresearch/macpine
cd macpine
make all
```