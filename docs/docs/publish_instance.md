# Publishing and Importing Instances

In order to back up or share instances (including `macpine` configurations and the VM filesystem), `alpine publish` and
`alpine import` are provided.

## Basic Usage

```sh
$ alpine list
NAME       OS         STATUS      SSH    PORTS ARCH        PID
example    alpine     Stopped     22           aarch64     -
$ alpine publish example
...
```

`publish` will then create `example.tar.gz`, a compressed archive of the relevant `macpine` and VM files. Note that `publish`
will stop a running instance before creating an archive.

```sh
$ alpine import example.tar.gz
```

`import` will then create a VM `example` (or `example-1`, `example-2`, ... if the name is taken).

## Publishing with Authenticated Encryption

`macpine` supports using [`age`](https://github.com/FiloSottile/age), a modern file encryption tool, to add authenticated encryption
to archives. `age` supports encryption with passphrases, but also with SSH keys (symmetrically with private keys, or asymmetrically
with public keys) and its own `age-keygen` custom format [`Curve25519`](https://en.wikipedia.org/wiki/Curve25519) keys.

`age` can be installed with: `go install filippo.io/age/cmd/...@latest`, or with many common package managers including:
* macOS `brew` `port`
* *nix `apk` `pacman` `apt` `dnf` `emerge` `nix-env` `zypper` `xbps-install` `pkg` `pkg_add`
* Windows `choco` `scoop`

`alpine publish -h`:
```sh
Publish an instance.

Usage:
  alpine publish NAME

Flags:
  -h, --help                 help for publish
  -p, --password             Encrypt published VM with interactive passphrase prompt (symmetric).
  -s, --private-key string   Encrypt published VM with ssh/age secret key (symmetric).
  -k, --public-key string    Encrypt published VM with ssh/age public key (asymmetric).
```

`-p` encrypts the archive `example.tar.gz.age` with a key derived from a securely-generated (default) or user-provided passphrase.
This option will prompt the user interactively for a passphrase input, and with no input (just `[Return]`) will generate and display
a secure passphrase. This is symmetric encryption: the same passphrase is used to encrypt and decrypt (when using `import`).

`-s` encrypts the archive *symmetrically* with a key derived from the provided *secret* key file. The same key file must be provided
for decryption during `import`.

`-k` encrypts the archive *asymmetrically* with a key derived from the provided *public* key file. The corresponding *secret* key
(a.k.a. private key) must be provided for decryption.

## Importing an Encrypted Archive

`age` uses authenticated encryption, therefore any modification to the data of a VM archive will cause decryption (and therefore `import`)
to fail with an error. Note that if the password/secret key is lost and the original unencrypted archive deleted, data may be
irretrievably lost.

If `age` is used with a password (`-p`/`--password` during `publish`), `age` will store this information in the archive. Therefore,
no command flag is needed -- `age` will detect and interactively prompt for the passphrase. If a public (`-k`) or private (`-s`)
key is used to encrypt, the (corresponding or same, respectively) private key must be supplied with the `-s` flag during `import`.

```sh
$ ssh-keygen -t ed25519
...
$ alpine list
NAME       OS         STATUS      SSH    PORTS ARCH        PID
example    alpine     Stopped     22           aarch64     -
$ alpine export -k ~/.ssh/id_ed25519.pub example # encrypt using public key
$ alpine delete example
$ alpine list
NAME       OS         STATUS      SSH    PORTS ARCH        PID
$ alpine import -s ~/.ssh/id_ed25519 example     # decrypt using private key
$ alpine list
NAME       OS         STATUS      SSH    PORTS ARCH        PID
example    alpine     Stopped     22           aarch64     -
```

Encrypting a VM archive to an SSH key of a GitHub user `username`:
```sh
$ curl https://github.com/username.keys -o username_key
$ alpine publish -k username_key example-vm
```
