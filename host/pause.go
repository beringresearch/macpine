package host

import (
	"github.com/beringresearch/macpine/qemu"
)

// Stop launches a new VM using user-defined configuration
func Pause(config qemu.MachineConfig) error {
	return config.Pause()
}

func Resume(config qemu.MachineConfig) error {
	return config.Resume()
}
