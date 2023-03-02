# How to create an instance



## Host-to-VM Port Forwarding

Network ingress over the virtual interface can be enabled during VM creation or after a "reboot" (`alpine stop ... ; alpine start ...`).

When using `-p` in `alpine launch` or adding to the `port` string in `config.yaml` (with `alpine edit` or otherwise), a string argument
must be provided. This string identifies ports which should be forwarded from the host to the guest by QEMU. A single port number will
forward that port on the host to that port on the guest, and a colon-delimited pair specifies mapping from host to guest with differing
port numbers.

The string can be described formally in pseudo-EBNF:

```
ports := "" | <port>,<ports>
port := number[u] | number:number[u]
number := 0 to 65535
```

Or informally as a `,` comma-delimited list of zero or more port mappings. A port mapping is either a number between 0 and 65536,
or two such port numbers separated by a `:` colon. An optional character `u` can be appended to configure a UDP port forward.

For example, to forward port 8080 from host to guest: `-p 8080` or `port: "8080"`

Further examples:

TCP forward `host:1111` to `guest:1111` and `host:2222` to `guest:3333`; UDP forward `host:4444` to `guest:4444` and `host:5555` to
`guest:6666`.

```
port: "1111,2222:3333,4444u,5555:6666u"
```

Forward 8080 from host to guest on TCP and UDP:

```
port: "8080,8080u"
```

## Configuring SSH and Storing SSH Credentials

By default, `macpine` requires `root` ssh to access and execute commands on guest machines. The default credential is the root password,
which is set (insecurely) to `root`. In most cases, this is sufficient for the use cases `macpine` is expected to support, as security
against malicious host system behavior is not within the threat model.

However, more secure credentials such as certificate-based ssh, VM hardening (e.g. disabling password-based ssh), or security best
practices may require credentials to be changed from the default, and stored outside the host filesystem.

In order to support multiple credential methods, `macpine` supports multiple credential "backends":

* `raw` i.e. password-based ssh, with password stored in `config.yaml`, default `root`
* `env` i.e. password-based ssh, with password stored in a host-system environment variable
* `ssh` i.e. [`ssh-agent`](https://www.ssh.com/academy/ssh/agent)-based ssh authentication

The second, `env`, is marginally more secure than `raw` and may be useful in automation scenarios or when `ssh-agent` is not available.
The third defers credential management to the host system's `ssh-agent`, which can be backed by hardened memory-based storage (default)
or credential managers such as `gnome-keyring-daemon` or the macOS system keychain.

In order to configure credentials in `config.yaml` for `sshpassword` (and `rootpassword` if `sshuser` is changed from the default of
`root`), credential strings describe to `macpine` how to authenticate via ssh. Credential strings take the following forms:

```
sshpassword: "password" # with sshuser: "root", authenticate via password
OR
sshpassword: "raw::password" # equivalent, prefix denotes the "raw" credential backend
OR
sshpassword: "env::SOME_VARIABLE" # ssh password is stored in environment variable $SOME_VARIABLE on the host
OR
sshpassword: "ssh::HOSTNAME" # ssh credential is stored in ssh-agent, and is configured for use with host HOSTNAME (e.g. in ~/.ssh/config)
```

If the `ssh` backend is used, ssh must be configured (usually in `~/.ssh/config`) with the given hostname to use the appropriate
credential, likely an [ssh private key](https://www.redhat.com/sysadmin/key-based-authentication-ssh).

For example, with keypair `id_ed25519` and `id_ed25519.pub`:

`~/.ssh/config`:
```
Host alpine
    Hostname localhost
    User root
    Port 22
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes
```

`~/.macpine/vm-name/config.yaml`:
```
alias: vm-name
image: alpine_3.16.0-aarch64.qcow2
arch: aarch64
cpu: "4"
memory: "2048"
disk: 10G
mount: ""
port: ""
sshport: "22"
sshuser: root
sshpassword: "ssh::alpine"
macaddress: 00:11:22:33:44:55
location: /Users/username/.macpine/vm-name
```

and ("inside" the VM) `/root/.ssh/authorized_keys`:
```
... contents of id_ed25519.pub ...
```
