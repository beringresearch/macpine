package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/spf13/cobra"
)

// stopCmd stops an Alpine instance
var stopCmd = &cobra.Command{
	Use:   "stop NAME",
	Short: "Stop an Alpine VM.",
	Run:   stop,
}

func stop(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing name - please provide VM name")
		return
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
