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
	"gopkg.in/yaml.v3"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <instance> [<instance>...]",
	Short: "Delete named instances.",
	Run:   delete,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func delete(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		log.Fatal("missing VM name")
	}

	vmList := host.ListVMNames()
	for _, vmName := range args {
		exists := utils.StringSliceContains(vmList, vmName)
		if !exists {
			log.Fatal("unknown machine " + vmName)
		}

		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		machineConfig := qemu.MachineConfig{}

		config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", vmName, "config.yaml"))
		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(config, &machineConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = host.Stop(machineConfig)
		if err != nil {
			log.Println("error stopping vm: " + err.Error())
		}
		os.RemoveAll(machineConfig.Location)
	}
}
