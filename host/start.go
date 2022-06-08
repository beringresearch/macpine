package host

import (
	"errors"
	"strings"

	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
)

// Start launches a new VM using user-defined configuration
func Start(config qemu.MachineConfig) error {

	status, _ := config.Status()
	if status == "Running" {
		return errors.New(config.Alias + " is already running")
	}

	ports := strings.Split(config.Port, ",")
	allPorts := append([]string{config.SSHPort}, ports...)

	for _, p := range allPorts {
		err := utils.Ping("localhost", p)
		if err != nil {
			return err
		}
	}

	err := config.Start()
	if err != nil {
		return err
	}

	return nil
}
