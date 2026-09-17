## Troubleshooting

### Common issue remediation/prevention

To solve some common issues:

* Ensure `PermitRootLogin yes` remains set in `/etc/ssh/sshd_config` (in the instance) or the machine may become inaccessible/fail to start.
* If a custom root password (e.g. `pass`) is set (in the instance), add `rootpassword: pass` in `config.yaml` via `alpine edit machine-name`
  or directly with any text editor.
* If `alpine list` reports a machine is `Running` but the process has been terminated/killed, deleting the PID file at
  `~/.macpine/machine-name/alpine.pid` may resolve the issue. `killall qemu-system` may also be useful to hard stop any running instances if needed.

### Adjusting time

Time sync issues between the host and a VM are [well known](https://github.com/canonical/multipass/issues/2430). For example, when the
host is suspended, the VM clock will also stop ticking.

To re-adjust a `macpine` instance real-time clock to its system clock, execute (inside the instance):

```bash
hwclock -s
```

Or on the host:

```bash
alpine exec instance-name "hwclock -s"
```

Also, consider an `ntp` daemon within your instance to maintain the system clock. This can be added inside your instance:

```bash
apk update; apk add openntpd
rc-update add openntpd default
rc-service openntpd start
```
or
```bash
apk update; apk add chrony
service chronyd start
```

More information on `chronyd` can be found on [the Arch wiki](https://wiki.archlinux.org/title/Chrony)

### Networking issues

* Due to [how `qemu` forwards network connections](https://wiki.qemu.org/Documentation/Networking#User_Networking_(SLIRP)) from the guest out via the host, utilities such as `ping` may not work (as ICMP is not handled).
* If an instance fails to start with a port error, there may be a listener already bound to the requested port(s). Ensure that the `ssh` port and any ports on the host side in the `Ports` configuration are mutually exclusive between instances which must run simultaneously.
* `netstat -anp tcp` and `netstat -anp udp` can be used to discover active `LISTEN` connections on the host. Ensure no other running services have bound ports that are configured to be forwarded to an instance (`ssh` or otherwise).
* `qemu` binds `0.0.0.0` for forwarded ports. This means that by default any source IP may send traffic to a guest. If the host system
    does not have a [firewall enabled](https://support.apple.com/guide/mac-help/change-firewall-settings-on-mac-mh11783/mac) then any
    machines which can reach the host can send traffic to the guest. If this is not desired, enable a host firewall. You do not need to
    click "Allow" for incoming connection to `qemu` when prompted by macOS as loopback connections (i.e. directly from the host itself)
    will still be allowed.

### Instance gets a DHCP address but can't be reached (VPN / network extension filtering)

If `alpine start`/`ssh`/`exec` resolves a `vmnet` IP address for the instance (or hangs for a long time on "getting instance IP address from DHCP leases" before eventually printing a hint about this), but the connection still never succeeds even though the guest itself finished booting and started `sshd` — the most likely cause is a VPN client or endpoint-security tool's **network extension** filtering or tunneling local network traffic on a per-app basis.

This can affect one terminal app's processes (e.g. iTerm2) while leaving others (a different app, or the instance's own guest-side traffic) completely unaffected, which makes it look like a `macpine`/networking bug rather than a host-level one. A key symptom: a raw connection test run from a *different* process/app succeeds instantly, while the same test from your actual terminal fails with `No route to host`.

**Important**: uninstalling the offending app (FortiClient, and similar corporate VPN/EDR tools, are common culprits) does **not** remove its System Extension — those are separate, persistent, kernel-level components. To fully deactivate one:

1. Check what's active: `systemextensionsctl list`
2. Remove any relevant entries (VPN/network filter extensions) via **System Settings → General → Login Items & Extensions → Network Extensions**.
3. Reboot — network extensions can keep their kernel-level hooks alive until the network stack reinitializes, even after being deactivated in System Settings.

Verify with `systemextensionsctl list` again after rebooting to confirm the extension is actually gone before retrying.

### Instance hangs on boot with no console output (`qemu` 11.1.1 regression)

On Apple Silicon, `aarch64` instances using `vmnet` networking (`vmnet: true` in `config.yaml`) may fail to boot when using `qemu` 11.1.1: the `qemu-system-aarch64` process pins a CPU core near 100%, produces no serial console output at all, and the instance never acquires a DHCP lease or becomes reachable. This reproduces with a bare `qemu-system-aarch64` invocation (outside of `macpine`), so it is not a `macpine` bug — it appears to be a regression in `qemu` 11.1.1 itself affecting early boot/firmware on the `aarch64` `virt` machine type with HVF acceleration.

Downgrading to `qemu` 10.0.3 resolves the issue. If you have an older `10.0.3` keg still available via Homebrew:

```bash
brew unlink qemu
brew link qemu@10.0.3   # or manually symlink the qemu-system-* binaries from
                         # /opt/homebrew/Cellar/qemu/10.0.3/bin into /opt/homebrew/bin
brew pin qemu            # prevent `brew upgrade` from reintroducing the regression
```

If Homebrew has already removed the old keg, you can also build/install `qemu` 10.0.3 from source or an older bottle. Track upstream for a fix before unpinning.

### Other issues

* If alpine is not able to resize the disk, it will error out with this message: `unable to resize disk: signal: abort trap`. Internally, it runs the command `qemu-img resize <IMAGE_LOCATION> <+SIZE>`. If the `qemu-img resize` command errors out with `dyld[...]: Library not loaded: /opt/homebrew/opt/libunistring/lib/libunistring.2.dylib` then re-installing `gettext` via `brew reinstall gettext` may resolve the issue.
