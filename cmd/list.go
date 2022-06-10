package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// listCmd lists Alpine instances
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available instances.",
	Run:   list,
}

func list(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	status := []string{}
	config := []qemu.MachineConfig{}
	pid := []int{}

	vmNames, err := host.ListVMNames()
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range vmNames {

		machineConfig := qemu.MachineConfig{}
		c, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", f, "config.yaml"))
		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(c, &machineConfig)
		if err != nil {
			log.Fatal(err)
		}

		config = append(config, machineConfig)

		s, p := host.Status(machineConfig)
		status = append(status, s)
		pid = append(pid, p)
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAME\tOS\tSTATUS\tSSH\tPORTS\tARCH\tPID\t")
	for i, machine := range config {
		fmt.Fprintln(w, machine.Alias+"    \t"+strings.Split(machine.Image, "_")[0]+"    \t"+status[i]+"    \t"+machine.SSHPort+"    \t"+machine.Port+"    \t"+machine.Arch+"    \t"+fmt.Sprint(pid[i])+"    \t")
	}
	w.Flush()

}
