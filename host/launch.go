package host

import (
	"errors"

	"github.com/beringresearch/macpine/qemu"
)

// Launch launches a new VM using user-defined configuration
func Launch(config qemu.MachineConfig) error {
	err := config.Launch()
	if err != nil {
		return err
	}

	err = config.Start()
	if err != nil {
		return errors.New("unable launch a new machine. " + err.Error())
	}

	return nil
}
