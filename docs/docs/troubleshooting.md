## Troubleshooting

To solve some common issues:

* Ensure `PermitRootLogin yes` remains set in `/etc/ssh/sshd_config` (in the VM) or the machine may become inaccessible/fail to start.
* If a custom root password (e.g. `pass`) is set (in the VM), add `rootpassword: pass` in `config.yaml` via `alpine edit machine-name`
  or directly with any text editor.
* If `alpine list` reports a machine is `Running` but the process has been terminated/killed, deleting the PID file at
  `~/.macpine/machine-name/alpine.pid` may resolve the issue.
