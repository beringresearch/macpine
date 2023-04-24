# How to modify an instance

## Adjusting time

Time sync issues between the host and a VM are [well known](https://github.com/canonical/multipass/issues/2430). For example, when the host is suspended, VM clock will also stop ticking.

To re-adjust a `macpine` instance real-time clock to its system clock, execute (inside the instance):

```bash
hwclock -s
```

Also, consider an `ntp` daemon such as `chrony` within your instance to maintain the system clock.

## Changing configurations

When modifying an instance with `alpine edit <instance name>` or directly modifying the `config.yaml` file,
`alpine restart <instance name>` will be required for changes to take effect.

Do not modify the `alias` entry in `config.yaml`, rather, use `alpine rename <instance name> <new name>` to rename instances.
