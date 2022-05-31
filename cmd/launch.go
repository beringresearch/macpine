package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// launchCmd launches an Alpine instance
var launchCmd = &cobra.Command{
	Use:   "launch FLAGS",
	Short: "Launch an Alpine VM.",
	Run:   launch,
}

var machineArch, machineCPU, machineMemory, machineDisk, machinePort string

func init() {
	includeLaunchFlags(launchCmd)
}

func includeLaunchFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&machineArch, "arch", "a", "x86-64", "Machine architecture. default: x86-64")
	cmd.Flags().StringVarP(&machineCPU, "cpu", "c", "1", "Number of CPUs to allocate. Minimum: 1, default: 1.")
	cmd.Flags().StringVarP(&machineMemory, "memory", "m", "2048", "Amount of memory to allocate. Positive integers, in bytes. Minimum: 128, default: 2048.")
	cmd.Flags().StringVarP(&machineDisk, "disk", "d", "10G", "Disk space to allocate. Positive integers, in bytes, or with K, M, G suffix. Minimum: 512M, default: 10G.")
	cmd.Flags().StringVarP(&machinePort, "port", "p", "10022", "Make VM accessible via SSH on this port. default: 10022")
}

func launch(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	machineConfig := qemu.MachineConfig{
		Alias:   utils.GenerateRandomAlias(),
		Arch:    machineArch,
		Version: "3.16.0",
		CPU:     machineCPU,
		Memory:  machineMemory,
		Disk:    machineDisk,
		Port:    machinePort,
	}
	machineConfig.Location = filepath.Join(userHomeDir, ".macpine", machineConfig.Alias)

	err = host.Launch(machineConfig)
	if err != nil {
		log.Fatal(err)
	}

}
