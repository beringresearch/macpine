package host

import (
	"github.com/beringresearch/macpine/qemu"
)

// Stop launches a new VM using user-defined configuration
func Stop(config qemu.MachineConfig) error {
	err := config.Stop()
	if err != nil {
		return err
	}

	return nil
}
