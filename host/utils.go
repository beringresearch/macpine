package host

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/beringresearch/macpine/qemu"
	"github.com/spf13/cobra"
)

func ListVMNames() []string {
	var vmList []string

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return vmList
	}

	dirList, err := os.ReadDir(filepath.Join(userHomeDir, ".macpine"))
	if err != nil {
		return vmList
	}

	for _, f := range dirList {
		if f.Name() != "cache" {
			vmList = append(vmList, f.Name())
		}
	}

	return vmList
}

// autocomplete with VM Names
func AutoCompleteVMNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return ListVMNames(), cobra.ShellCompDirectiveNoFileComp
}

func AutoCompleteVMNamesOrTags(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	vmNames := ListVMNames()
	tags, err := ListTags()
	if err != nil {
		tags = []string{}
	}
	for i, t := range tags {
		tags[i] = "+" + t
	}
	return append(vmNames, tags...), cobra.ShellCompDirectiveNoFileComp
}

func ListTags() ([]string, error) {
	var tagList []string

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dirList, err := os.ReadDir(filepath.Join(userHomeDir, ".macpine"))
	if err != nil {
		return nil, err
	}

	for _, f := range dirList {
		if f.Name() != "cache" {
			machineConfig, err := qemu.GetMachineConfig(f.Name())
			if err != nil {
				return nil, err
			}
			tagList = append(tagList, machineConfig.Tags...)
		}
	}
	return tagList, nil
}

func ExpandTagArguments(args []string) ([]string, error) {
	var expandedArgs []string
	var tags []string
	var tagMap = make(map[string]([]string))

	for _, arg := range args {
		if strings.HasPrefix(arg, "+") {
			tags = append(tags, arg[1:])
		}
	}
	if len(tags) == 0 {
		return args, nil
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dirList, err := os.ReadDir(filepath.Join(userHomeDir, ".macpine"))
	if err != nil {
		return nil, err
	}

	for _, f := range dirList {
		if f.Name() != "cache" {
			machineConfig, err := qemu.GetMachineConfig(f.Name())
			if err != nil {
				return nil, err
			}
			for _, tag := range machineConfig.Tags {
				if arr, ok := tagMap[tag]; ok {
					tagMap[tag] = append(arr, f.Name())
				} else {
					tagMap[tag] = []string{f.Name()}
				}
			}
		}
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "+") {
			if vms, ok := tagMap[arg[1:]]; ok {
				expandedArgs = append(expandedArgs, vms...)
			} else {
				return nil, errors.New("no instances found with tag " + arg[1:])
			}
		} else {
			expandedArgs = append(expandedArgs, arg)
		}
	}

	return expandedArgs, nil
}
