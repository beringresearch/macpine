package cmd

import (
	"errors"
	"log"
	"os"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <instance> [<instance>...]",
	Short:   "Delete instances.",
	Run:     delete,
	Aliases: []string{"del", "rm", "remove"},

	ValidArgsFunction:     host.AutoCompleteVMNamesOrTags,
	DisableFlagsInUseLine: true,
}

func delete(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		log.Fatal("missing instance name")
	}

	args, err := host.ExpandTagArguments(args)
	if err != nil {
		log.Fatalln(err)
	}

	vmList := host.ListVMNames()
	errs := make([]utils.CmdResult, len(args))
	for i, vmName := range args {
		if utils.StringSliceContains(args[:i], vmName) {
			continue
		}
		exists := utils.StringSliceContains(vmList, vmName)
		if !exists {
			errs[i] = utils.CmdResult{Name: vmName, Err: errors.New("unknown instance " + vmName)}
			continue
		}

		machineConfig, err := qemu.GetMachineConfig(vmName)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		err = host.Stop(machineConfig)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		err = os.RemoveAll(machineConfig.Location)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}
		log.Printf("instance %s deleted\n", vmName)
	}
	wasErr := false
	for _, res := range errs {
		if res.Err != nil {
			log.Printf("failed to delete %s: %v\n", res.Name, res.Err)
			wasErr = true
		}
	}
	if wasErr {
		log.Fatalln("error deleting instance(s)")
	}
}
