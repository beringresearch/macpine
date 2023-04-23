package cmd

import (
	"errors"
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

var deleteCmd = &cobra.Command{
	Use:   "delete <instance> [<instance>...]",
	Short: "Delete instances.",
	Run:   delete,

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

		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		machineConfig := qemu.MachineConfig{}

		config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", vmName, "config.yaml"))
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		err = yaml.Unmarshal(config, &machineConfig)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		err = host.Stop(machineConfig)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		os.RemoveAll(machineConfig.Location)
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
