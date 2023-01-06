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

// stopCmd stops an Alpine instance
var stopCmd = &cobra.Command{
	Use:   "stop NAME",
	Short: "Stop an instance.",
	Run:   stop,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func stop(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing VM name")
	}

	vmList, err := host.ListVMNames()
	if err != nil {
		log.Fatal(err)
	}

	exists := utils.StringSliceContains(vmList, args[0])
	if !exists {
		log.Fatal("unknown machine " + args[0])
	}

	machineConfig := qemu.MachineConfig{
		Alias: args[0],
	}
	machineConfig.Location = filepath.Join(userHomeDir, ".macpine", machineConfig.Alias)

	err = host.Stop(machineConfig)
	if err != nil {
		log.Fatal(err)
	}

}
