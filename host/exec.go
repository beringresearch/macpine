package host

import (
	"github.com/beringresearch/macpine/qemu"
)

// Exec executes a command inside VM
func Exec(config qemu.MachineConfig, cmd string) error {
	return config.Exec(cmd, false) // false: run as default ssh user, not (necessarily) root
}
