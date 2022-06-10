package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// startCmd starts an Alpine instance
var startCmd = &cobra.Command{
	Use:   "start NAME",
	Short: "Start an instance.",
	Run:   start,
}

func start(cmd *cobra.Command, args []string) {

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

	machineConfig := qemu.MachineConfig{}

	config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", args[0], "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = host.Start(machineConfig)
	if err != nil {
		log.Fatal(err)
	}

}
