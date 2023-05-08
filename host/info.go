package host

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/qemu"
	"gopkg.in/yaml.v3"
)

func Info(vmName string) (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", vmName, "config.yaml"))
	if err != nil {
		return "", err
	}

	machineConfig := qemu.MachineConfig{}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		return "", err
	}

	info := fmt.Sprintf("Name: %s\nImage: %s\nArch: %s\nDisk size: %s\nMemory size: %s\nCPUs: %s\nMount: %s\nTags: %s\n",
		machineConfig.Alias,
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
