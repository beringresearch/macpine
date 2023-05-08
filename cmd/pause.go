package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// stopCmd stops an Alpine instance
var pauseCmd = &cobra.Command{
	Use:   "pause <instance> [<instance>...]",
	Short: "Pause instances.",
	Run:   pause,

	ValidArgsFunction:     host.AutoCompleteVMNamesOrTags,
	DisableFlagsInUseLine: true,
}

func pause(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing instance name")
	}

	args, err = host.ExpandTagArguments(args)
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

		machineConfig := qemu.MachineConfig{
			Alias: vmName,
		}
		machineConfig.Location = filepath.Join(userHomeDir, ".macpine", machineConfig.Alias)

		err = host.Pause(machineConfig)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}
	}
	wasErr := false
	for _, res := range errs {
		if res.Err != nil {
			log.Printf("failed: %v\n", res.Err)
			wasErr = true
		}
	}
	if wasErr {
		log.Fatalln("error pausing instance(s)")
	}
}
