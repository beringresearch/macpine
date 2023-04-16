# Hardening `macpine` instances (via QEMU use and configuration)

### QEMU Resources

* [General QEMU Security Information](https://www.qemu.org/docs/master/system/security.html)
  * `!!` note: tiny codegen emulation is not developed for security. Guests emulated using `tcg` must be considered trusted.
* [Reporting QEMU Security Issues](https://www.qemu.org/contribute/security-process/)

### Linux Resources

* [grsecurity](https://grsecurity.net) (non-free)
* [SELinux](https://selinuxproject.org/page/Main_Page) and [AppArmor](https://ubuntu.com/server/docs/security-apparmor)

### General tips

* Limit exposed interfaces, virtual devices, and ports
* Run services on unprivileged ports (> 1024) as
  [dedicated users](https://security.stackexchange.com/questions/47576/do-simple-linux-servers-really-need-a-non-root-user-for-security-reasons)
  with localhost proxying if needed
* Configure `ssh-agent` authentication to the guest machine with certificate-based credentials, and then disable password
  authentication (`PermitRootLogin prohibit-password` and/or `PasswordAuthentication no` in `/etc/ssh/sshd_config`)
* `qemu` port forwarding binds `0.0.0.0`, meaning any source IP may send traffic to the guest. Enabling a firewall on the host can prevent
    unwanted ingress traffic to the guest.
