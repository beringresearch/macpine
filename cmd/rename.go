package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var renameCmd = &cobra.Command{
	Use:   "rename <instance> <name>",
	Short: "Rename an instance.",
	Run:   rename,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func rename(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) < 1 {
		log.Fatalln("missing VM name")
	}
	if len(args) < 2 {
		log.Fatalln("missing new name argument")
	}

	vmName := args[0]
	vmList := host.ListVMNames()
	exists := utils.StringSliceContains(vmList, vmName)
	if !exists {
		log.Fatalln("unknown machine " + vmName)
	}

	newName := args[1]
	validateName(newName)

	configDir := filepath.Join(userHomeDir, ".macpine")
	files, err := os.ReadDir(configDir)
	if err != nil {
		log.Fatalf("error reading macpine config directory: %v\n", err)
	}

	for _, oldVM := range files {
		if newName == oldVM.Name() {
			log.Fatalln("cannot rename: name is already taken")
		}
	}

	machineConfig := qemu.MachineConfig{}

	config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", vmName, "config.yaml"))
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(config, &machineConfig)
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

	config, err = yaml.Marshal(&machineConfig)
	if err != nil {
		log.Fatalf("error serializing config: %v\n", err)
	}

	err = ioutil.WriteFile(filepath.Join(newLocation, "config.yaml"), config, 0644)
	if err != nil {
		log.Fatalf("error writing updated config: %v\n", err)
	}

	log.Printf("renamed '%s' to '%s'\n", vmName, newName)
}

func validateName(name string) {
	if name == "cache" {
		log.Fatalln("cannot rename: 'cache' is reserved")
	}
	if strings.HasPrefix(name, ".") {
		log.Fatalln("cannot rename: name must not begin with '.'")
	}
	format := regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	if !format.MatchString(name) {
		log.Fatalln("cannot rename: invalid name, accepted characters are [A-Za-z0-9], '.', '_', and '-'")
	}
}
