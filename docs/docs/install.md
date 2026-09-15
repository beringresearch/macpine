# Installation

`macpine` depends on QEMU >= 6.2.0_1:

```bash
#brew update
#brew upgrade
brew install qemu
```

> **Known issue:** QEMU 11.1.1 has a regression that can hang `aarch64` instances on boot when using `vmnet` networking. See [Troubleshooting](troubleshooting.md#instance-hangs-on-boot-with-no-console-output-qemu-1111-regression) if you hit this.

## Running `vmnet` networking without `sudo`

Instances with `vmnet: true` in `config.yaml` (bridged networking, giving the instance a real LAN-reachable IP) normally require `alpine start`/`stop` to run as root. That's because QEMU's built-in `vmnet-shared` networking calls macOS's `Vmnet.framework` directly, which requires root privileges (or a special Apple-granted entitlement that a plain Homebrew build of QEMU doesn't have).

To avoid `sudo` for every `vmnet` VM start/stop, install [`socket_vmnet`](https://github.com/lima-vm/socket_vmnet) — the same privileged-helper daemon used by Lima and Colima. It does the one privileged `vmnet` setup step in the background; unprivileged processes (including `macpine`) then talk to it over a Unix socket instead of calling `Vmnet.framework` themselves:

```bash
brew install socket_vmnet
sudo brew services start socket_vmnet   # one-time; runs the daemon persistently in the background
```

Once the daemon is running, `macpine` detects it automatically and uses it for any `vmnet` instance — no configuration needed, and no more `sudo` on `alpine start`/`stop`/`ssh`/`exec` for those instances. If `socket_vmnet` isn't installed or running, `macpine` falls back to QEMU's native `vmnet-shared` networking (requiring `sudo`) exactly as before.

## Install the latest binary

Download the [latest binary release](https://github.com/beringresearch/macpine/releases) for your system and add it to your path by placing to e.g. `/usr/local/bin/`

```bash
wget https://github.com/beringresearch/macpine/releases/download/v1.0/alpine_darwin_arm64
mv alpine_darwin_arm64 alpine
sudo chmod +x alpine
sudo mv alpine /usr/local/bin/
```

## Install via Homebrew (recommended)

```bash
brew install macpine
```

## Install via MacPorts
On MacOS, you can install via [MacPorts](https://www.macports.org/):

```bash
sudo port install macpine
```

## Install from source

```bash
git clone https://github.com/beringresearch/macpine
cd macpine
make
make install #install to /usr/local by default, may require sudo
```
