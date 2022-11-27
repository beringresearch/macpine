# Running LXD containers in Macpine

## Overview

[LXD](https://linuxcontainers.org/lxd/introduction/) is a next generation system container manager with support for [a wide number of Linux distributions](https://uk.lxd.images.canonical.com). It provides a simple way to build, test, and run multiple Linux environments across a single machine or multiple compute clusters.

Under the hood, LXD uses [LXC](https://linuxcontainers.org/lxc/introduction/), through liblxc and its Go binding, to create and manage the containers. However, LXD relies on a number of Linux kernel features, such as CGroups and kernel namespaces, which aren't natively available on MacOS.

Macpine makes it possible to run LXD/LXC containers on MacOS with support for both amd64 and arm64 processors, through its lightweight virtualisation layer. This workflow makes it easy to develop and test LXD containers locally.

## Prerequisites

1. Install QEMU

```bash
brew install qemu
```

2. Install [the latest Macpine binary](https://github.com/beringresearch/macpine#install-the-latest-binary)

3. Install the LXD client

```bash
brew install lxc
```

## Launch an LXD VM

Now that that the system is ready, we can create a lightweight Macpine VM, which has been preconfigured to run LXD. In your terminal run:

```bash
alpine launch --image alpine_3.16.0_lxd --name lxd --port 8443
```

This will create a new VM called `lxd` and forward port `8443` (the default port that LXD client uses to communicate with the LXD server) of the VM to host. Macpine will attempt to match the native CPU architecture of your host to the correct VM image. However, if you can explicitly specify the architecture by adding either `--arch aarch64` or `--arch x86_64` to the above command.

## Configure LXD

Before you can create an instance, you need to configure LXD.

Run the following command to start the interactive configuration process:

```bash
alpine exec lxd "lxd init"
```

For the purposes of this tutorial, it is recommended to accept default settings.

>> NOTE: the above command is executed inside your `lxd` VM and is sandboxed from your host.

## Configure LXD remote

Set up your LXD remote to communicate with the LXD client on your host.

```bash
alpine exec lxd "lxc config set core.https_address 0.0.0.0"
alpine exec lxd "lxc config set core.trust_password root"
```

>> NOTE: for the purposes of this demonstration, the remote password is configured as `root`.

## Add the remote to your LXD host:

```bash
lxc remote add macpine 127.0.0.1
```

Accept the certificate and type `root` for Admin password (this password can be configured with `lxc config set core.trust_password` above).

Finally, set this remote as the default:

```bash
lxc remote switch macpine
```

That's it - you can now run LXD containers through Macpine at nearly-native speeds!

## Launching your first LXD container

LXD containers can now be launched and manipulated through the `lxc` client:

```bash
lxc launch ubuntu
```