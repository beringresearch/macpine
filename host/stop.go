package host

import (
	"github.com/beringresearch/macpine/qemu"
)

// Stop launches a new VM using user-defined configuration
func Stop(config qemu.MachineConfig) error {
	return config.Stop()
}
