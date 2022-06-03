package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// stopCmd stops an Alpine instance
var execCmd = &cobra.Command{
	Use:   "exec NAME COMMAND",
	Short: "execute COMMAND on an Alpine VM.",
	Run:   exec,
}

func exec(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatal("missing VM name")
		return
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