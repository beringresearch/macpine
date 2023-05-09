package cmd

import (
	"errors"
	"log"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// startCmd starts an Alpine instance
var startCmd = &cobra.Command{
	Use:     "start <instance> [<instance>...]",
	Short:   "Start instances.",
	Run:     start,
	Aliases: []string{"boot", "on"},

	ValidArgsFunction:     host.AutoCompleteVMNamesOrTags,
	DisableFlagsInUseLine: true,
}

func start(cmd *cobra.Command, args []string) {
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

		if status, _ := machineConfig.Status(); status != "Stopped" {
			errs[i] = utils.CmdResult{Name: vmName, Err: errors.New(vmName + " is already running")}
			continue
		}
		err = host.Start(machineConfig)
		if err != nil {
			host.Stop(machineConfig)
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}
	}
	wasErr := false
	for _, res := range errs {
		if res.Err != nil {
			log.Printf("failed to start %s: %v\n", res.Name, res.Err)
			wasErr = true
		}
	}
	if wasErr {
		log.Fatalln("error starting instance(s)")
	}
}
