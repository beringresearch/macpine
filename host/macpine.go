package host

import (
	"fmt"

	"github.com/beringresearch/macpine/config"
)

// Launch launches a new VM using user-defined configuration
func Launch(config config.MachineConfig) error {
	fmt.Println(config)

	return nil
}
