package host

import (
	"log"
	"strconv"
	"strings"

	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
)

// Start launches a new VM using user-defined configuration
func Start(config qemu.MachineConfig) error {

	status, _ := config.Status()
	if status == "Running" {
		log.Println(config.Alias + " is already running")
		return nil
	}

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
		if strings.Contains(p, ":") {
			p = strings.Split(p, ":")[0]
		}
		err := utils.Ping("localhost", p)
		if err != nil {
			return err
		}
	}

	err = config.Start()
	if err != nil {
		return err
	}

	return nil
}
