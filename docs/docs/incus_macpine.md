# Running LXD containers in Macpine

## Overview

[Incus](https://linuxcontainers.org/incus/) is a next generation system container manager with support for [a wide number of Linux distributions](https://images.linuxcontainers.org). It provides a simple way to build, test, and run multiple Linux environments across a single machine or multiple compute clusters.

Under the hood, LXD uses [LXC](https://linuxcontainers.org/lxc/introduction/), through liblxc and its Go binding, to create and manage the containers. However, incus relies on a number of Linux kernel features, such as CGroups and kernel namespaces, which aren't natively available on MacOS.

[Macpine](https://github.com/beringresearch/macpine) makes it possible to run Incus/LXC containers on MacOS with support for both amd64 and arm64 processors, through its lightweight virtualisation layer. This workflow makes it easy to develop and test incus containers locally.

## Prerequisites

1. Install QEMU and macpine

```bash
brew install qemu macpine
```

2. Install the incus client

```bash
brew install incus
```

## Install Incus
Now that that the system is ready, we can create a lightweight Macpine instance, which will be configured to run Incus. In your terminal run:

```bash
alpine launch --name incus --ssh 223 --port 8443
```

This will create a new instance called `incus` and forward port `8443` (the default port that LXD client uses to communicate with the LXD server) of the instance to host. Macpine will attempt to match the native CPU architecture of your host to the correct instance image. However, if you can explicitly specify the architecture by adding either `--arch aarch64` or `--arch x86_64` to the above command.

Now lets install the incus daemon.

```bash
alpine exec incus -- "hwclock -s; wget https://raw.githubusercontent.com/beringresearch/macpine/main/scripts/enable_incus.sh"

alpine exec incus -- "ash enable_incus.sh"
```

When the script finishes execution, the incus daemon will be available at guest startup.

## Configure incus
Before you can create an instance, you need to configure incus.

Run the following command to accept all automatic defaults:

```bash
alpine exec incus "incus admin init --auto"
```

For the purposes of this tutorial, it is recommended to accept default settings.

>> NOTE: the above command is executed inside your `incus` instance and is sandboxed from your host.

## Configure incus remote

Set up your incus remote to communicate with the incus client on your host.

```bash
alpine exec incus "incus config set core.https_address 0.0.0.0:8443"
alpine exec incus "incus config trust add mymac"
```

The command generates and prints a token that can be used to add the client certificate.

>> NOTE: Make a note of the token as it will be used to authenticate the incus client.

## Add the remote to your Incus client:

Now that the remote server is configured, lets finish by configuring the incus client and adding our `incus` macpine instance as a remote.

```bash
incus remote add incus https://127.0.0.1
```

Enter the trust token for incus that you've noted from the steps before.

Finally, set this remote as the default:

```bash
incus remote switch incus
```

That's it - you can now run incus containers through Macpine at nearly-native speeds!

## Launching your first incus container

Incus containers can now be launched and manipulated through the `incus` client. On your mac run:

```bash
incus launch images:debian/bullseye debian
```


## Connecting to your first LXD container

```bash
incus exec debian -- bash
```

## Saving VM status

You can save the Macpine VM image with your incus configuration for later use:

```bash
alpine publish incus
```

This will create a tar ball that can be imported using `alpine import`.

## Cleanup

```bash
incus stop debian
incus delete debian
alpine stop incus
alpine delete incus
```