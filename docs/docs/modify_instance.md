# How to modify an instance

## Adjusting time

Timesync issues between the host and a VM are [well known](https://github.com/canonical/multipass/issues/2430). For example, when the host is suspended, VM clock will also stop ticking.

To re-adjust guest VM to host clock, execute inside your VM:

```bash
hwclock -s
```