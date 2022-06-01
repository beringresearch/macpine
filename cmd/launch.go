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

var machineArch, machineVersion, machineCPU, machineMemory, machineDisk, machinePort string

func init() {
	includeLaunchFlags(launchCmd)
}

func includeLaunchFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&machineVersion, "version", "v", "3.16.0", "Alpine image version. default: 3.16.0")
	cmd.Flags().StringVarP(&machineArch, "arch", "a", "x86_64", "Machine architecture. default: x86_64")
	cmd.Flags().StringVarP(&machineCPU, "cpu", "c", "4", "Number of CPUs to allocate. Minimum: 1, default: 1.")
	cmd.Flags().StringVarP(&machineMemory, "memory", "m", "2048", "Amount of memory to allocate. Positive integers, in bytes. Minimum: 128, default: 2048.")
	cmd.Flags().StringVarP(&machineDisk, "disk", "d", "10G", "Disk space to allocate. Positive integers, in bytes, or with K, M, G suffix. Minimum: 512M, default: 10G.")
	cmd.Flags().StringVarP(&machinePort, "port", "p", "22", "Forward VM ports to host. Multiple ports can be separate by `,`. default: 22")
}

func launch(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	machineConfig := qemu.MachineConfig{
		Alias:   utils.GenerateRandomAlias(),
		Image:   "alpine_" + machineVersion + "-" + machineArch + ".qcow2",
		Arch:    machineArch,
		Version: machineVersion,
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
