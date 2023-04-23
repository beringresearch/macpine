package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// execCmd executes command on alpine vm
var execCmd = &cobra.Command{
	Use:   "exec <instance> <command>",
	Short: "execute a command on an instance over ssh.",
	Run:   exec,

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

	instName := args[0]
	cmdArgs := strings.Join(args[1:], " ")

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	machineConfig := qemu.MachineConfig{}

	config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", instName, "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = host.Exec(machineConfig, cmdArgs)
	if err != nil {
		log.Fatal(err)
	}
}
