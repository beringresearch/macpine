package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// shellCmd starts an Alpine instance
var shellCmd = &cobra.Command{
	Use:   "ssh <instance>",
	Short: "Attach an interactive shell to an instance via ssh.",
	Run:   shell,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func shell(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing instance name")
	}

	vmList := host.ListVMNames()
	exists := utils.StringSliceContains(vmList, args[0])
	if !exists {
		log.Fatal("unknown instance " + args[0])
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

	if status, _ := machineConfig.Status(); status != "Running" {
		log.Fatalf("%s is not running", machineConfig.Alias)
	}
	err = host.Exec(machineConfig, "bash")
	if err != nil {
		log.Fatal(err)
	}
}
