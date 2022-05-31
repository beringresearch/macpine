package host

import (
	"github.com/beringresearch/macpine/qemu"
)

// Start launches a new VM using user-defined configuration
func Start(config qemu.MachineConfig) error {
	err := config.Start()
	if err != nil {
		return err
	}

	return nil
}
