package host

import (
	"strconv"

	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
)

// Launch launches a new VM using user-defined configuration
func Launch(config qemu.MachineConfig) error {

	ports, err := utils.ParsePort(config.Port)
	if err != nil {
		return err
	}
	hostports := make([]string, len(ports))
	for i, p := range ports {
		hostports[i] = strconv.Itoa(p.Host)
	}
	allPorts := append([]string{config.SSHPort}, hostports...)

	for _, p := range allPorts {
		err := utils.Ping("localhost", p)
		if err != nil {
			return err
		}
	}

	err = config.Launch()
	if err != nil {
		config.Stop()
		config.CleanPIDFile()
		return err
	}

	return nil
}
