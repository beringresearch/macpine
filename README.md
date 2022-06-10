[![Go Report Card](https://goreportcard.com/badge/github.com/beringresearch/macpine)](https://goreportcard.com/report/github.com/beringresearch/macpine)

# Lightweight Alpine VMs on MacOS

Create and manage lightweight Alpine VMs on MacOS with:


:repeat: Seamless port forwarding

:card_index_dividers: Automatic file sharing

:roller_coaster: Bridged networking

:rocket: aarch64 and x86_64 emulation

## Motivation
The goal of this project is to enable MacOS users to:

1. Easily spin up and manage lightweight Alpine Linux environments.
2. Use tiny VMs to take advantage of containerisation technologies, including [LXD](https://linuxcontainers.org/lxd/introduction/) and Docker.
3. Build and test software on x86_64 and aarch64 systems.

# Installation

Macpine depends on QEMU >= 6.2.0_1:

```bash
#brew update
#brew upgrade
brew install qemu
```

## Install the latest binary

Download the [latest binary release](https://github.com/beringresearch/macpine/releases) for your system and add it to your path by placing to e.g. `/usr/local/bin/`

```bash
wget https://github.com/beringresearch/macpine/releases/download/v.05/alpine
sudo mv alpine /usr/local/bin/
```

## Install from source

```bash
git clone https://github.com/beringresearch/macpine
cd macpine
make darwin
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
alpine launch -s 22 #launch a VM and expose SSH port to host port 22
ssh root@localhost -p 22 #password: root
```

Expose additional VM ports to host:

```bash
alpine launch -s 23 -p 8888,5432 #launch a VM, expose SSH to host port 23 and forward VM ports 8888 and 5432 to host ports 8888 and 5432
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

## Resizing disk partitions

The easiest way to resize a VM disk is:

1. Stop the target VM: `alpine stop $VMNAME`
2. Back up VM disk with  `cp ~/.macpine/$VMNAME/$VMIMAGENAME backup.qcow2`
3. Adjust disk size with `qemu-img resize ~/.macpine/$VMNAME/$VMIMAGENAME +20G`
4. Start the target VM:  `alpine start $VMNAME`
5. Adjust guest disk partition size: `cfdisk`
6. Tell the OS that disk has been expanded: `resize2fs /dev/vda*`

## Command Reference

```bash
alpine --help
Create, control and connect to Alpine instances.

Usage:
  alpine [command]

Available Commands:
  delete      Delete an instance.
  edit        Edit instance configuration using Vim.
  exec        execute COMMAND over ssh.
  help        Help about any command
  import      Imports an instance.
  launch      Launch an Alpine instance.
  list        List all available instances.
  publish     Publish an instance.
  ssh         Attach an interactive shell to an instance.
  start       Start an instance.
  stop        Stop an instance.

Flags:
  -h, --help   help for alpine

Use "alpine [command] --help" for more information about a command.
```