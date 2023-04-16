## Troubleshooting

### Common issue remediation/prevention

To solve some common issues:

* Ensure `PermitRootLogin yes` remains set in `/etc/ssh/sshd_config` (in the VM) or the machine may become inaccessible/fail to start.
* If a custom root password (e.g. `pass`) is set (in the VM), add `rootpassword: pass` in `config.yaml` via `alpine edit machine-name`
  or directly with any text editor.
* If `alpine list` reports a machine is `Running` but the process has been terminated/killed, deleting the PID file at
  `~/.macpine/machine-name/alpine.pid` may resolve the issue. `killall qemu-system` may also be useful to hard stop any running VMs if needed.

### Networking issues

* Due to [how `qemu` forwards network connections](https://wiki.qemu.org/Documentation/Networking#User_Networking_(SLIRP)) from the guest out via the host, utilities such as `ping` may not work (as ICMP is not handled).
* If a VM fails to start with a port error, there may be a listener already bound to the requested port(s). Ensure that the `ssh` port and any ports on the host side in the `Ports` configuration are mutually exclusive between VMs which must run simultaneously.
* `netstat -anp tcp` and `netstat -anp udp` can be used to discover active `LISTEN` connections on the host. Ensure no other running services have bound ports that are configured to be forwarded to a VM (`ssh` or otherwise).
* `qemu` binds `0.0.0.0` for forwarded ports. This means that by default any source IP may send traffic to a guest. If the host system
    does not have a [firewall enabled](https://support.apple.com/guide/mac-help/change-firewall-settings-on-mac-mh11783/mac) then any
    machines which can reach the host can send traffic to the guest. If this is not desired, enable a host firewall. You do not need to
    click "Allow" for incoming connection to qemu when prompted by macOS as loopback connections (i.e. directly from the host itself) will
    still be allowed.
