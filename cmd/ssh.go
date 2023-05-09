package cmd

import (
	"log"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
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
	if len(args) == 0 {
		log.Fatal("missing instance name")
	}

	vmName := args[0]
	vmList := host.ListVMNames()
	exists := utils.StringSliceContains(vmList, vmName)
	if !exists {
		log.Fatalln("unknown instance " + vmName)
	}

	machineConfig, err := qemu.GetMachineConfig(vmName)

	if status, _ := machineConfig.Status(); status != "Running" {
		log.Fatalf("%s is not running", machineConfig.Alias)
	}
	err = host.Exec(machineConfig, "bash")
	if err != nil {
		log.Fatalln(err)
	}
}
