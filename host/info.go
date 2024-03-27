package host

import (
	"fmt"

	"github.com/beringresearch/macpine/qemu"
)

func Info(vmName string) (string, error) {
	machineConfig, err := qemu.GetMachineConfig(vmName)
	if err != nil {
		return "", err
	}

	info := fmt.Sprintf("Name: %s\nIP: %s\nImage: %s\nArch: %s\nDisk size: %s\nMemory size: %s\nCPUs: %s\nMount: %s\nTags: %s\n",
		machineConfig.Alias,
		machineConfig.MachineIP,
		machineConfig.Image,
		machineConfig.Arch,
		machineConfig.Disk,
		machineConfig.Memory,
		machineConfig.CPU,
		machineConfig.Mount,
		machineConfig.Tags,
	)
	return info, nil
}
