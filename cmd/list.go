package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/spf13/cobra"
)

// listCmd lists Alpine instances
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List instances.",
	Run:     list,
	Aliases: []string{"ls"},

	DisableFlagsInUseLine: true,
}

func list(cmd *cobra.Command, args []string) {
	status := []string{}
	configs := []qemu.MachineConfig{}
	pid := []string{}

	vmNames := host.ListVMNames()
	for _, vmName := range vmNames {
		machineConfig, err := qemu.GetMachineConfig(vmName)
		if err != nil {
			log.Fatal(err)
		}

		configs = append(configs, machineConfig)

		s, p := host.Status(machineConfig)
		status = append(status, s)
		if s == "Stopped" {
			pid = append(pid, "-")
		} else {
			pid = append(pid, fmt.Sprint(p))
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tSSH\tPORTS\tARCH\tPID\tTAGS\t")
	for i, machine := range configs {
		spacer := "    \t"
		row := []string{
			machine.Alias,
			status[i],
			machine.SSHPort,
			machine.Port,
			machine.Arch,
			fmt.Sprint(pid[i]),
			strings.Join(machine.Tags, ","),
		}
		fmt.Fprintln(w, strings.Join(row, "    \t")+spacer)
	}
	w.Flush()
}
