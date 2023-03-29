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

Download the [latest binary release](https://github.com/beringresearch/macpine/releases) for your system and add it to your path by placing to e.g. `/usr/local/bin/`:

```bash
wget https://github.com/beringresearch/macpine/releases/download/v0.10/alpine_darwin_arm64
mv alpine_darwin_arm64 alpine
sudo chmod +x alpine
sudo mv alpine /usr/local/bin/
```

## Install from Homebrew

```bash
brew install macpine
```

## Install from MacPorts

On macOS, you can install via [MacPorts](https://www.macports.org):

```bash
sudo port install macpine
```

See more information [here](https://ports.macports.org/port/macpine/).

## Install from source

```bash
git clone https://github.com/beringresearch/macpine
cd macpine
make
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
alpine ssh $VMNAME #attach to the VM shell
```

Expose additional VM ports to host:

```bash
alpine launch -s 23 -p 8888,5432 #launch a VM, expose SSH to host port 23 and forward host ports 8888 and 5432 to VM ports 8888 and 5432
alpine launch -s 8023 -p 8081:8082,8083 #launch a VM, expose SSH to host port 8023, forward host port 8081 to VM port 8082, and forward
                                        #host port 8083 to VM port 8083
```

VMs can be easily packaged for export and re-use as tar.gz files:

```bash
alpine list

NAME                 OS         STATUS      SSH    PORTS ARCH      PID     TAGS
cheerful-result      alpine     Running     25           aarch64   26568
glittering-swing     alpine     Running     23           x86_64    57206   emulation,intel
```

```bash
alpine publish cheerful-result
```

This will create a file cheerful-result.tar.gz which can be imported as:

```bash
alpine import cheerful-result.tar.gz
```

## Command Reference

```bash
Create, control and connect to Alpine instances.

Usage:
  alpine [command]

Available Commands:
  completion  Generate shell autocompletions.
  delete      Delete named instances.
  edit        Edit instance configuration.
  exec        execute COMMAND over ssh.
  help        Help about any command
  import      Imports an instance.
  info        Display information about instances.
  launch      Launch an Alpine instance.
  list        List all available instances.
  publish     Publish an instance.
  ssh         Attach an interactive shell to an instance.
  start       Start an instance.
  stop        Stop an instance.
  tag         Add or remove tags from an instance.

Flags:
  -h, --help   help for alpine

Use "alpine [command] --help" for more information about a command.
```

Shell command completion files can be generated with `alpine completion [bash|zsh|fish|powershell]`.
See `alpine completion -h` or the [completion documentation](docs/docs/completions.md) for more information.
