package host

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func ListVMNames() []string {
	var vmList []string

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return vmList
	}

	dirList, err := ioutil.ReadDir(filepath.Join(userHomeDir, ".macpine"))
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
