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

`macpine` is intended for use on modern macOS. Support for older versions of macOS and other OSes may vary.

## Install from Homebrew (recommended)

```bash
brew install macpine # installs `alpine` command and `qemu` dependency automatically
```

## Get the latest binary

Download the [latest binary release](https://github.com/beringresearch/macpine/releases/latest) for your system and add it to your `$PATH`
by moving it to e.g. `/usr/local/bin/`:

```bash
arch="$([ `uname -m` = 'x86_64' ] && echo 'amd64' || echo 'arm64')" # detect architecture
wget "https://github.com/beringresearch/macpine/releases/latest/download/alpine_darwin_$arch"
mv "alpine_darwin_$arch" alpine
sudo chmod +x alpine
sudo mv alpine /usr/local/bin/
#export PATH="$PATH:/usr/local/bin"
```

Macpine depends on QEMU >= 6.2.0_1:

```bash
brew install qemu
```

## Install from MacPorts

You can also install `macpine` via [MacPorts](https://www.macports.org):

```bash
sudo port install macpine
```
## Install from source

Building from source requires a working `go` compiler, and running requires `qemu`:

```bash
brew install go qemu
git clone https://github.com/beringresearch/macpine
cd macpine
make            # compiles the project into a local bin/ directory
make install    # installs binaries to /usr/local/bin
                # PREFIX=/some/other/path make install installs to /some/other/path
```

# Getting Started

To create and start a new instance:

```bash
alpine launch                       # launches with default parameters
alpine launch -a aarch64            # create an aarch64 instance
alpine launch -d 10G -c 4 -m 2048   # create an instance with a 10GB disk, 4 cpus, and 2GB of RAM

alpine launch -h                    # view all configuration options and defaults
```

Access instance via ssh:

```bash
alpine launch -s 22         # launch an instance and expose SSH port to host port 22
alpine ssh instance-name    # attach shell to instance (replace `instance-name` as appropriate)
```

Expose additional instance ports to host:

```bash
# launch an instance, expose SSH to host port 2022 and forward host ports 8888 and 5432 to instance ports 8888 and 5432
alpine launch -s 2022 -p 8888,5432

# launch an instance, expose SSH to host port 8022, forward host port 8081 to instance port 8082, and forward
# host port 8083 to instance port 8083
alpine launch -s 8022 -p 8081:8082,8083

# launch an instance, expose SSH to host port 9022, forward host port 9091 UDP to instance port 9091 UDP,
# and forward host port 9092 UDP to instance port 9093 UDP
alpine launch -s 9023 -p 9091u,9092:9093u
```

Instances can be easily packaged for backup or sharing as `.tar.gz` files:

```bash
alpine list

NAME                 STATUS      SSH    PORTS   ARCH      PID     TAGS
cheerful-result      Running     2022           aarch64   26568
glittering-swing     Running     3022           x86_64    57206   emulation,intel
```

```bash
alpine publish cheerful-result
```

This will create a file `cheerful-result.tar.gz` which can be imported as:

```bash
#alpine delete cheerful-result
alpine import cheerful-result.tar.gz
```

See [all the docs](docs/docs) for more information:
- [advanced port forwarding and securing instance SSH](docs/docs/create_instance.md)
- [editing instance configurations](docs/docs/modify_instance.md)
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
  edit        Edit instance configurations.
  exec        execute a command on an instance over ssh.
  help        Help about any command
  import      Imports an instance archived with `alpine publish`.
  info        Display information about instances.
  launch      Create and start an instance.
  list        List instances.
  pause       Pause instances.
  publish     Publish instances.
  rename      Rename an instance.
  restart     Stop and start instances.
  resume      Unpause instances.
  ssh         Attach an interactive shell to an instance via ssh.
  start       Start instances.
  stop        Stop instances.
  tag         Add or remove tags from an instance.

Flags:
  -h, --help   help for alpine

Use "alpine [command] --help" for more information about a command.
```

**Multiple instances in a command:** some commands (`delete`, `edit`, `info`, `pause`, `publish`, `restart`, `resume`, `start`, `stop`)
accept multiple instance names and will repeat the operation over each unique named instance once.

**Tags:** using `alpine tag`, instances can be tagged; tags can be used in multi-instance commands (see above) e.g.
`alpine start +foobar` will start all instances which have had been tagged `foobar` with `alpine tag instance-name foobar`. Note that
the tag `launchctl-autostart` [is used for auto-starting instances at login](docs/docs/autostart.md).

**Shell autocompletion:** shell command completion files (installed automatically with `brew install macpine`) can be generated with
`alpine completion [bash|zsh|fish|powershell]`. See `alpine completion -h` or the [completion documentation](docs/docs/completions.md)
for more information.
