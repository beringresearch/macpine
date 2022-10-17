package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// importCmd iports an Alpine VM from file
var importCmd = &cobra.Command{
	Use:   "import NAME",
	Short: "Imports an instance.",
	Run:   importMachine,
}

func importMachine(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		log.Fatal("missing VM name")
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	_, err = utils.CopyFile(args[0], filepath.Join(userHomeDir, ".macpine", args[0]))
	if err != nil {
		log.Fatal(err)
	}

	targetDir := filepath.Join(userHomeDir, ".macpine", strings.Split(args[0], ".tar.gz")[0])
	err = utils.Uncompress(filepath.Join(userHomeDir, ".macpine", args[0]), targetDir)
	if err != nil {
		log.Fatal(err)
	}

	machineConfig := qemu.MachineConfig{}
	config, err := ioutil.ReadFile(filepath.Join(targetDir, "config.yaml"))
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal(err)
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal(err)
	}

	machineConfig.Alias = strings.Split(args[0], ".tar.gz")[0]
	machineConfig.Location = targetDir

	updatedConfig, err := yaml.Marshal(&machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(targetDir, "config.yaml"), updatedConfig, 0644)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal(err)
	}

	if err != nil {
		err = os.Remove(filepath.Join(userHomeDir, ".macpine", args[0]))
		if err != nil {
			log.Fatal("unable to import: " + err.Error())
		}

		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}

	err = os.Remove(filepath.Join(userHomeDir, ".macpine", args[0]))
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}
}
