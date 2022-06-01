package host

import (
	"errors"
	"net"

	"github.com/beringresearch/macpine/qemu"
)

// Start launches a new VM using user-defined configuration
func Start(config qemu.MachineConfig) error {
	ln, err := net.Listen("tcp", ":"+config.Port)

	if err != nil {
		return errors.New("can't listen on port " + config.Port + ": " + err.Error())
	}

	_ = ln.Close()

	err = config.Start()
	if err != nil {
		return err
	}

	return nil
}
