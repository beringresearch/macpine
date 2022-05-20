package host

import (
	"errors"
	"log"

	"github.com/beringresearch/macpine/vm"
)

// Launch launches a new VM using user-defined configuration
func Launch(config vm.MachineConfig) error {

	log.Printf("Launching " + config.Alias)

	err := config.Init()
	if err != nil {
		return err
	}

	err = config.CreateQemuDiskImage("user-data.qcow2")
	if err != nil {
		return errors.New("unable to create disk image. " + err.Error())
	}

	err = config.CreateQemuDiskImage("tmp.qcow2")
	if err != nil {
		return errors.New("unable to create disk image. " + err.Error())
	}

	err = config.Install()
	if err != nil {
		return errors.New("unable launch a new machine. " + err.Error())
	}

	return nil
}
