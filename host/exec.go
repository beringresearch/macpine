package host

import (
	"github.com/beringresearch/macpine/qemu"
)

// Exec executes a command inside VM
func Exec(config qemu.MachineConfig, cmd string) error {

	if cmd != "ash" {
		err := config.Exec(cmd)
		if err != nil {
			return err
		}
	} else {
		err := config.ExecShell()
		if err != nil {
			return err
		}
	}

	return nil
}
