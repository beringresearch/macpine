# Running LXD containers in Macpine

## Overview

[LXD](https://linuxcontainers.org/lxd/introduction/) is a next generation system container manager with support for [a wide number of Linux distributions](https://uk.lxd.images.canonical.com). It provides a simple way to build, test, and run multiple Linux environments across a single machine or multiple compute clusters.

Under the hood, LXD uses [LXC](https://linuxcontainers.org/lxc/introduction/), through liblxc and its Go binding, to create and manage the containers. However, LXD relies on a number of Linux kernel features, such as CGroups and kernel namespaces, which aren't natively available on MacOS.

Macpine makes it possible to run LXD/LXC containers on MacOS with support for both amd64 and arm64 processors, through its lightweight virtualisation layer. This workflow makes it easy to develop and test LXD containers locally.

## Prerequisites

1. Install QEMU and macpine

```bash
brew install qemu macpine
```

2. Install the LXD client

```bash
brew install lxc
```

## Launch an LXD instance

Now that that the system is ready, we can create a lightweight Macpine instance. In your terminal run:

```bash
sudo alpine launch --name lxd-aarch64 --shared
```

This will create a new instance called `lxd-aarch64`.

## Configure LXD

Before you can create an instance, you need to configure LXD.

Run the following command to accept all automatic defaults:

```bash
alpine exec lxd-aarch64 "wget https://raw.githubusercontent.com/beringresearch/macpine/refs/heads/main/scripts/enable_lxd.sh"
alpine exec lxd-aarch64 "ash enable_lxd.sh"
alpine exec lxd-aarch64 "lxd init --auto"
```

For the purposes of this tutorial, it is recommended to accept default settings.

>> NOTE: the above command is executed inside your `lxd-aarch64` instance and is sandboxed from your host.

## Configure LXD remote

Set up your LXD remote to communicate with the LXD client on your host.

```bash
alpine exec lxd-aarch64 "lxc config set core.https_address [machineip]"
alpine exec lxd-aarch64 "lxc config set core.trust_password root"
```

Your VM's IP address is obtained by running `alpine info lxd-aarch64`.

>> NOTE: for the purposes of this demonstration, the remote password is configured as `root`. This password can be configured with `lxc config set core.trust_password` above

## Add the remote to your LXD host:

```bash
lxc remote add macpine [machineip] --accept-certificate --password root
```

Your VM's IP address is obtained by running `alpine info lxd-aarch64`.

>> NOTE: if you create an alpine lxd instance, then destroy it, then try to reconfigure another on later on your host, you may need to delete `macpine` remote from `~/.config/lxc/config.yml` due to new certificates each time.

Finally, set this remote as the default:

```bash
lxc remote switch macpine
```

That's it - you can now run LXD containers through Macpine at nearly-native speeds!

## Launching your first LXD container

LXD containers can now be launched and manipulated through the `lxc` client:

```bash
lxc launch ubuntu:24.04 ubuntu
```

## Mounting host directory -> lxd Macpine instance -> lxd container

```bash
lxc config device add ubuntu share disk source=/root/mnt path=/root/mnt
```

## Connecting to your first LXD container

```bash
lxc exec ubuntu -- bash
```

## Cleanup

```bash
lxc stop ubuntu
lxc delete debian
alpine delete lxd-aarch64
```
