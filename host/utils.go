package host

import (
	"io/ioutil"
	"os"
	"path/filepath"
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
