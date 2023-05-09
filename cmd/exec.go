package cmd

import (
	"log"
	"strings"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// execCmd executes command on alpine vm
var execCmd = &cobra.Command{
	Use:     "exec <instance> <command>",
	Short:   "execute a command on an instance over ssh.",
	Run:     exec,
	Aliases: []string{"x", "execute", "cmd", "command"},

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func exec(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatal("missing instance name")
	}

	vmList := host.ListVMNames()
	exists := utils.StringSliceContains(vmList, args[0])
	if !exists {
		log.Fatal("unknown instance " + args[0])
	}

	vmName := args[0]
	cmdArgs := strings.Join(args[1:], " ")

	machineConfig, err := qemu.GetMachineConfig(vmName)
	if err != nil {
		log.Fatalln(err)
	}

	err = host.Exec(machineConfig, cmdArgs)
	if err != nil {
		log.Fatalln(err)
	}
}
