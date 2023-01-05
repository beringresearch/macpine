package host

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func ListVMNames() ([]string, error) {
	var vmList []string

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return vmList, nil
	}

	dirList, err := ioutil.ReadDir(filepath.Join(userHomeDir, ".macpine"))
	if err != nil {
		return vmList, nil
	}

	for _, f := range dirList {
		if f.Name() != "cache" {
			vmList = append(vmList, f.Name())
		}

	}

	return vmList, nil
}

// autocomplete with VM Names
func AutoCompleteVMNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	listVMNames, err := ListVMNames()
	if err != nil {
		log.Fatal(err)
	}

	return listVMNames, cobra.ShellCompDirectiveNoFileComp
}
