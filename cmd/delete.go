package cmd

import (
	"errors"
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
	Use:   "delete NAME",
	Short: "Delete an instance.",
	Run:   delete,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func delete(cmd *cobra.Command, args []string) {

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

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
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

	err = host.Stop(machineConfig)
	if err != nil {
		log.Fatal(errors.New("unable to stop VM: " + err.Error()))
	}

	os.RemoveAll(machineConfig.Location)

}
