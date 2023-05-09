# Editing instance configurations

Various instance parameters are set at creation with `alpine launch`. These can be changed at any time, but if an instance is
running, a reboot (`alpine restart instance-name`) is required for them to take effect.

## Modifying instance configs

`alpine edit instance-name` will open the configuration file in a terminal editor, `$EDITOR`, `vim`, or `nano` by default.
Configuration files can be found in `~/.macpine/instance-name/config.yaml` for editing with external tools.

Do not modify the `alias` or `location` entries in `config.yaml`, rather, use `alpine rename <instance name> <new name>` to rename
instances.

Some validations are performed after an `alpine edit` editing, and if they fail the `config.yaml` will be reverted to its pre-edit state.

## Config file format

The instance configurations are stored as [`YAML`](https://yaml.org) in their respective instance directories in `~/.macpine`.
An illustrative example is shown here with comments:

```yaml
alias: instance-name                            # instance name for use in `alpine` commands, only modify with `alpine rename`
image: alpine_3.16.0-aarch64.qcow2              # image file in ~/.macpine/cache to boot from
arch: aarch64                                   # architecture, either ARM or Intel
cpu: "2"                                        # number of virtual threads to allocate
memory: "2048"                                  # megabytes (mebibytes, really) of RAM to allocate
disk: 10G                                       # bytes of storage to allocate
mount: "/Users/user/Documents"                  # directories to mount to /mnt in the instance
port: "8080,9090u,10010:10020"                  # port forwarding specification (refer to `docs/docs/create_instance.md`)
sshport: "20022"                                # host port for SSH, forwards to TCP/22 on the instance
sshuser: root                                   # can be modified, but then `rootpassword` must be specified
sshpassword: root                               # can be hardened with other authentication (refer to `docs/docs/create_instance.md`)
rootpassword: pass                              # optional, only required if `sshuser` is changed from `root`
macaddress: aa:bb:cc:dd:ee:ff                   # generated, no need to modify
location: /Users/user/.macpine/instance-name    # location on host filesystem, only modify with `alpine rename`
tags:                                           # instance tags in `alpine list` and `alpine <command> +foo` tag-based commands
    - foo
    - bar
    - baz
```
