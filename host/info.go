package host

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/qemu"
	"gopkg.in/yaml.v3"
)

func Info(vmName string) (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", vmName, "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	machineConfig := qemu.MachineConfig{}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		log.Fatal(err)
	}

	info := fmt.Sprintf("Name: %s\nArch: %s\nDisk usage: %s\nMemory usage: %s\nCPU usage: %s\nMounts: %s",

		machineConfig.Alias,
		machineConfig.Arch,
		machineConfig.Disk,
		machineConfig.Memory,
		machineConfig.CPU,
		machineConfig.Mount)

	return info, nil
}
