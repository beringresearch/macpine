# Auto-start instances at login

`macpine` can add a `launchctl` agent to start instances upon user login. Note that this may slow the login process.

Once installed, the launch agent will automatically start all instances with the tag `launchctl-autostart` upon user login. This tag can be added
with `alpine tag <instance name> launchctl-autostart` to an arbitrary number of instances.

There are two ways to install the `macpine` `launchctl` launch agent `plist` file:

## Installing with `brew` and `brew services`

```bash
brew install macpine
brew services start macpine
```

## Installing from source

```bash
git clone https://github.com/beringresearch/macpine.git
cd macpine
make install && make agent
```

This will add `alpineDaemonLaunchAgent.plist` to `~/Library/LaunchAgents` with the directive to start the
appropriately tagged instances.
