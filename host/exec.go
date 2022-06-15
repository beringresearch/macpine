package host

import (
	"github.com/beringresearch/macpine/qemu"
)

// Exec executes a command inside VM
func Exec(config qemu.MachineConfig, cmd string) error {

	err := config.Exec(cmd)
	if err != nil {
		return err
	}

	return nil
}
