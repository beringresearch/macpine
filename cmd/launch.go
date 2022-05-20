package cmd

import (
	"log"

	"github.com/beringresearch/macpine/config"
	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// launchCmd launches an Alpine instance
var launchCmd = &cobra.Command{
	Use:   "launch FLAGS",
	Short: "Launch an Alpine VM.",
	Run:   launch,
}

var machineCPU, machineMemory, machineDisk, machinePort string

func init() {
	includeLaunchFlags(launchCmd)
}

func includeLaunchFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&machineCPU, "cpu", "c", "1", "Number of CPUs to allocate. Minimum: 1, default: 1.")
	cmd.Flags().StringVarP(&machineMemory, "memory", "m", "1GB", "Amount of memory to allocate. Positive integers, in bytes, or with K, M, G suffix. Minimum: 128M, default: 1G.")
	cmd.Flags().StringVarP(&machineDisk, "disk", "d", "5GB", "Disk space to allocate. Positive integers, in bytes, or with K, M, G suffix. Minimum: 512M, default: 5G.")
	cmd.Flags().StringVarP(&machinePort, "port", "p", "10022", "Make VM accessible via SSH on this port. default: 10022")
}

func launch(cmd *cobra.Command, args []string) {
	machineConfig := config.MachineConfig{
		Alias:  utils.GenerateRandomAlias(),
		CPU:    machineCPU,
		Memory: machineMemory,
		Disk:   machineDisk,
		Port:   machinePort,
	}

	err := host.Launch(machineConfig)
	if err != nil {
		log.Fatal(err)
	}

}
