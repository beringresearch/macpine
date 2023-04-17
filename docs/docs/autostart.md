# Auto-start instances at login

`macpine` can add a `launchctl` agent to start instances upon user login. Note that this may slow the login process.

The launch agent will automatically start all instances with the tag `launchctl-autostart`. This tag can be added
with `alpine tag <instance name> launchctl-autostart` to an arbitrary number of instances.

Installing the agent requires building `macpine` from source:

```bash
brew install macpine
git clone https://github.com/beringresearch/macpine.git
cd macpine
make agent
```

This will add `alpineDaemonLaunchAgent.plist` to `~/Library/LaunchAgents` with the directive to start the
appropriately tagged instances.
