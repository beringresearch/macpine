package cmd

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var deleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete an Alpine VM by NAME.",
	Run:   delete,
}

func delete(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing name - please provide VM name")
		return
	}

	machineConfig := qemu.MachineConfig{}

	config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", args[0], "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = host.Stop(machineConfig)
	if err != nil {
		log.Fatal(errors.New("unable to stop VM: " + err.Error()))
	}

	os.RemoveAll(machineConfig.Location)

}
