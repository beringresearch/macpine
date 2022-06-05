package host

import (
	"errors"
	"strings"

	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
)

// Launch launches a new VM using user-defined configuration
func Launch(config qemu.MachineConfig) error {

	ports := strings.Split(config.Port, ",")
	allPorts := append([]string{config.SSHPort}, ports...)

	for _, p := range allPorts {
		err := utils.Ping("localhost", p)
		if err != nil {
			return err
		}
	}

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
