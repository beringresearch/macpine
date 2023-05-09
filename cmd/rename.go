package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:     "rename <instance> <name>",
	Short:   "Rename an instance.",
	Run:     rename,
	Aliases: []string{"mv", "move"},

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func rename(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) < 1 {
		log.Fatalln("missing instance name")
	}
	if len(args) < 2 {
		log.Fatalln("missing new name argument")
	}

	vmName := args[0]
	vmList := host.ListVMNames()
	exists := utils.StringSliceContains(vmList, vmName)
	if !exists {
		log.Fatalln("unknown instance " + vmName)
	}

	newName := args[1]
	err = ValidateName(newName)
	if err != nil {
		log.Fatalln(err)
	}

	configDir := filepath.Join(userHomeDir, ".macpine")
	files, err := os.ReadDir(configDir)
	if err != nil {
		log.Fatalf("error reading config directory: %v\n", err)
	}

	for _, oldVM := range files {
		if newName == oldVM.Name() {
			log.Fatalln("cannot rename: name is already taken")
		}
	}

	machineConfig, err := qemu.GetMachineConfig(vmName)
	if err != nil {
		log.Fatalln(err)
	}

	oldLocation := machineConfig.Location
	newLocation := filepath.Join(configDir, newName)

	err = os.Rename(oldLocation, newLocation)
	if err != nil {
		log.Fatalf("error renaming config directory: %v\n", err)
	}

	machineConfig.Alias = newName
	machineConfig.Location = newLocation

	err = qemu.SaveMachineConfig(machineConfig)
	if err != nil {
		log.Fatalf("error writing updated config: %v\n", err)
	}

	log.Printf("renamed '%s' to '%s'\n", vmName, newName)
}

func ValidateName(name string) error {
	if name == "cache" {
		return errors.New("cannot rename: 'cache' is reserved")
	}
	if strings.HasPrefix(name, ".") {
		return errors.New("cannot rename: name must not begin with '.'")
	}
	format := regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	if !format.MatchString(name) {
		return errors.New("cannot rename: invalid name, accepted characters are [A-Za-z0-9], '.', '_', and '-'")
	}
	return nil
}
