[![build](https://github.com/beringresearch/macpine/actions/workflows/build.yml/badge.svg)](https://github.com/beringresearch/macpine/actions/workflows/build.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/beringresearch/macpine)](https://goreportcard.com/report/github.com/beringresearch/macpine)

# Lightweight Alpine VMs on MacOS

Create and manage lightweight Alpine VMs on MacOS with:

:repeat: Seamless port forwarding

:card_index_dividers: Automatic file sharing

:roller_coaster: Bridged networking

:rocket: aarch64 and x86_64 emulation

<img src="https://user-images.githubusercontent.com/5474655/233823903-357fcece-08e4-44fd-a173-b038c978de3f.gif" width="80%" />

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

## Install from Homebrew (recommended)

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
#brew install go
git clone https://github.com/beringresearch/macpine
cd macpine
make #compiles the project into a local bin/ directory
make install #installs binaries to /usr/local (or other configured PREFIX)
```

# Getting Started

To launch a new instance:

```bash
alpine launch #launches with default parameters
alpine launch -a aarch64 #create an aarch64 instance
alpine launch -d 10G -c 4 -m 2048 #create an instance with a 10GB disk, 4 cpus and 2GB of RAM
```

`alpine help launch` to view defaults and other launch options.

Access instance via ssh:

```bash
alpine launch -s 22 #launch an instance and expose SSH port to host port 22
alpine ssh <instance name> #attach shell to instance
#Or: ssh root@localhost -p 22 #password: root
```

Expose additional instance ports to host:

```bash
alpine launch -s 23 -p 8888,5432 #launch a VM, expose SSH to host port 23 and forward host ports 8888 and 5432 to VM ports 8888 and 5432
alpine launch -s 8023 -p 8081:8082,8083 #launch a VM, expose SSH to host port 8023, forward host port 8081 to VM port 8082, and forward
                                        #host port 8083 to VM port 8083
alpine launch -s 9023 -p 9091u,9092:9093u #launch a VM, expose SSH to host port 9023, forward (UDP) host port 9091 to VM port 9091,
                                          #and forward (UDP) host port 9092 to VM port 9093
```

Instances can be easily packaged for backup or sharing as tar.gz files:

```bash
alpine list

NAME                 STATUS      SSH    PORTS ARCH      PID     TAGS
cheerful-result      Running     25           aarch64   26568
glittering-swing     Running     23           x86_64    57206   emulation,intel
```

```bash
alpine publish cheerful-result
```

This will create a file `cheerful-result.tar.gz` which can be imported as:

```bash
alpine import cheerful-result.tar.gz
```

See [all the docs](docs/docs) for more information:
- [advanced port forwarding and securing instance SSH](docs/docs/create_instance.md)
- [managing instances](docs/docs/modify_instance.md)
- [running LXD within instances](docs/docs/lxd_macpine.md) 
- [hardening instances](docs/docs/hardening.md)
- [auto-starting instances at login](docs/docs/autostart.md)
- [general troubleshooting](docs/docs/troubleshooting.md)

## Command Reference

```man
Create, control, and connect to Alpine instances.

Usage:
  alpine [command]

Available Commands:
  completion  Generate shell autocompletions.
  delete      Delete instances.
  edit        Edit instance configuration.
  exec        execute a command on an instance over ssh.
  help        Help about any command
  import      Imports an instance archived with `alpine publish`.
  info        Display information about an instance.
  launch      Create and start an instance.
  list        List instances.
  pause       Pause an instance.
  publish     Publish an instance.
  rename      Rename an instance.
  restart     Stop and start an instance.
  resume      Unpause an instance.
  ssh         Attach an interactive shell to an instance via ssh.
  start       Start an instance.
  stop        Stop an instance.
  tag         Add or remove tags from an instance.

Flags:
  -h, --help   help for alpine

Use "alpine [command] --help" for more information about a command.
```

**Multiple instances in a command:** some commands (`delete`, `edit`, `publish`, `restart`, `start`, `stop`, `pause`, `resume`) accept multiple instance names and will repeat the operation over each unique named instance once.

**Tags:** using `alpine tag`, instances can be tagged; tags can be used in multi-instance commands (see above) e.g. `alpine start +daemon` will start all instances which have had been tagged `daemon` with `alpine tag <instance name> daemon`. Note that the tag `launchctl-autostart` [is used for auto-starting instances at login](docs/docs/autostart.md).

**Shell autocompletion:** shell command completion files (installed by default with `brew install macpine`) can be generated with `alpine completion [bash|zsh|fish|powershell]`.
See `alpine completion -h` or the [completion documentation](docs/docs/completions.md) for more information.
