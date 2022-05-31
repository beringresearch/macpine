package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// listCmd starts an Alpine instance
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available Alpine VM instances.",
	Run:   list,
}

func list(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	dirList, err := ioutil.ReadDir(filepath.Join(userHomeDir, ".macpine"))
	if err != nil {
		log.Fatal(err)
	}

	status := []string{}
	config := []qemu.MachineConfig{}
	pid := []int{}

	for _, f := range dirList {
		if f.Name() != "cache" {
			machineConfig := qemu.MachineConfig{}
			c, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", f.Name(), "config.yaml"))
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
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tPORT\tPID\t")
	for i, machine := range config {
		fmt.Fprintln(w, machine.Alias+"    \t"+status[i]+"    \t"+machine.Port+"    \t"+fmt.Sprint(pid[i])+"    \t")
	}
	w.Flush()

}
