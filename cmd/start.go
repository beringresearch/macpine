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

// startCmd starts an Alpine instance
var startCmd = &cobra.Command{
	Use:   "start <instance> [<instance>...]",
	Short: "Start an instance.",
	Run:   start,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func start(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing VM name")
	}

	vmList := host.ListVMNames()
	errs := make([]utils.CmdResult, len(args))
	for i, arg := range args {
		exists := utils.StringSliceContains(vmList, arg)
		if !exists {
			errs[i] = utils.CmdResult{Name: arg, Err: errors.New("unknown machine " + arg)}
			continue
		}

		machineConfig := qemu.MachineConfig{}

		config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", arg, "config.yaml"))
		if err != nil {
			errs[i] = utils.CmdResult{Name: arg, Err: err}
			continue
		}

		err = yaml.Unmarshal(config, &machineConfig)
		if err != nil {
			errs[i] = utils.CmdResult{Name: arg, Err: err}
			continue
		}

		err = host.Start(machineConfig)
		if err != nil {
			host.Stop(machineConfig)
			errs[i] = utils.CmdResult{Name: arg, Err: err}
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
		log.Fatalln("error starting VM(s)")
	}
}
