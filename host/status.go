package host

import (
	"github.com/beringresearch/macpine/qemu"
)

// Status launches a new VM using user-defined configuration
func Status(config qemu.MachineConfig) (string, int) {
	return config.Status() // status, pid
}
