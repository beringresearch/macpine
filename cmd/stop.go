package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// stopCmd stops an Alpine instance
var stopCmd = &cobra.Command{
	Use:   "stop <instance> [<instance>...]",
	Short: "Stop an instance.",
	Run:   stop,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func stop(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing VM name")
	}

	vmList := host.ListVMNames()
	errs := make([]utils.CmdResult, len(args))
	for i, vmName := range args {
      if utils.StringSliceContains(args[:i], vmName) {
         continue
      }
		exists := utils.StringSliceContains(vmList, vmName)
		if !exists {
			errs[i] = utils.CmdResult{Name: vmName, Err: errors.New("unknown machine " + vmName)}
			continue
		}

		machineConfig := qemu.MachineConfig{
			Alias: vmName,
		}
		machineConfig.Location = filepath.Join(userHomeDir, ".macpine", machineConfig.Alias)

		err = host.Stop(machineConfig)
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
		log.Fatalln("error stopping VM(s)")
	}
}
