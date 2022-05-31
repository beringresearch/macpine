package host

import (
	"errors"
	"log"

	"github.com/beringresearch/macpine/qemu"
)

// Launch launches a new VM using user-defined configuration
func Launch(config qemu.MachineConfig) error {

	log.Printf("Launching " + config.Alias)

	err := config.Init()
	if err != nil {
		return err
	}

	err = config.Install()
	if err != nil {
		return errors.New("unable launch a new machine. " + err.Error())
	}

	return nil
}
