package cmd

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// stopCmd stops an Alpine instance
var restartCmd = &cobra.Command{
	Use:   "restart <instance> [<instance>...]",
	Short: "Stop and start an instance.",
	Run:   restart,

	ValidArgsFunction:     host.AutoCompleteVMNamesOrTags,
	DisableFlagsInUseLine: true,
}

func restart(cmd *cobra.Command, args []string) {
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
	wasErr := false
	for i, vmName := range args {
		if utils.StringSliceContains(args[:i], vmName) {
			continue
		}
		exists := utils.StringSliceContains(vmList, vmName)
		if !exists {
			wasErr = true
			log.Println(errors.New("unknown instance " + vmName))
			continue
		}

		machineConfig := qemu.MachineConfig{}
		config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", vmName, "config.yaml"))
		if err != nil {
			wasErr = true
			log.Println(err)
			continue
		}
		err = yaml.Unmarshal(config, &machineConfig)
		if err != nil {
			wasErr = true
			log.Println(err)
			continue
		}

		log.Println("restarting " + vmName + "...")
		err = host.Stop(machineConfig)
		if err != nil {
			wasErr = true
			log.Println(err)
			continue
		}

		time.Sleep(time.Second)

		err = host.Start(machineConfig)
		if err != nil {
			host.Stop(machineConfig)
			wasErr = true
			log.Println(err)
			continue
		}
	}
	if wasErr {
		log.Fatalln("error restarting instance(s)")
	}
}
