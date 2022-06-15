# Running LXD containers in Macpine

## Overview

Macpine makes it possible to run LXD containers on MacOS with support for both amd64 and arm64 processors. This workflow makes it easy to develop and test LXD containers locally.

## Prerequisites

This tutorial assumes that Macpine and QEMU are installed on your system. Next, install the LXD client:

```bash
brew install lxd
```

## Launch an LXD VM

Now that Macpine is installed, we can create a VM running LXD. In your terminal run:

```bash
alpine launch --image alpine_3.16.0_lxd --name lxd --port 8443
```

This will create a new amd64 VM called LXD and forward port 8443 of the VM to host. Your LXD client will use port 8443 to connect to LXD server running inside the VM. If you require an arm64 VM, simply add `--arch aarch64` to the above command.

Configure LXD server inside the VM:

```bash
alpine exec lxd "lxd init"
```

For the purposes of this tutorial, it is recommended to accept default settings.

## Configure LXD remote

```bash
alpine exec lxd "lxc config set core.https_address 0.0.0.0"
alpine exec lxd "lxc config set core.trust_password root"
```

Now add the remote on your host:

```bash
lxc remote add macpine 127.0.0.1
```

Accept the certificate and type `root` for Admin password (this password can be configured with `lxc config set core.trust_password` above).

Finally, set this remote as the default:

```bash
lxc remote switch macpine
```

That's it - you can now run LXD containers through Macpine!

## Launching your first LXD container

LXD containers can now be launched and manipulated through the `lxc` client:

```bash
lxc launch ubuntu
```