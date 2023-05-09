package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/utils"
)

// infoCmd displays macpine machine info
var infoCmd = &cobra.Command{
	Use:     "info <instance> [<instance>...]",
	Short:   "Display information about instances.",
	Run:     macpineInfo,
	Aliases: []string{"i", "show"},

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func macpineInfo(cmd *cobra.Command, args []string) {
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
		info, err := host.Info(vmName)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}
		fmt.Print(info)
		if i < len(args)-1 {
			fmt.Println()
		}
	}
	wasErr := false
	for _, res := range errs {
		if res.Err != nil {
			log.Printf("error for %s: %v\n", res.Name, res.Err)
			wasErr = true
		}
	}
	if wasErr {
		log.Fatalln("error showing instance(s) info")
	}
}
