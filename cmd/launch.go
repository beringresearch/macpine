package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// launchCmd launches an Alpine instance
var launchCmd = &cobra.Command{
	Use:   "launch FLAGS",
	Short: "Launch an Alpine VM.",
	Run:   launch,
}

var machineArch, machineVersion, machineCPU, machineMemory, machineDisk, machinePort, sshPort, machineName, machineMount string

func init() {
	includeLaunchFlags(launchCmd)
}

func includeLaunchFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&machineVersion, "version", "v", "3.16.0", "Alpine image version.")
	cmd.Flags().StringVarP(&machineArch, "arch", "a", "x86_64", "Machine architecture.")
	cmd.Flags().StringVarP(&machineCPU, "cpu", "c", "4", "Number of CPUs to allocate.")
	cmd.Flags().StringVarP(&machineMemory, "memory", "m", "2048", "Amount of memory to allocate. Positive integers, in kilobytes.")
	cmd.Flags().StringVarP(&machineDisk, "disk", "d", "10G", "Disk space to allocate. Positive integers, in bytes, or with K, M, G suffix.")
	cmd.Flags().StringVarP(&machineMount, "mount", "", "", "Path to host directory to be exposed on guest. (default $HOME)")
	cmd.Flags().StringVarP(&sshPort, "ssh", "s", "22", "Forward VM SSH port to host.")
	cmd.Flags().StringVarP(&machinePort, "port", "p", "", "Forward VM ports to host. Multiple ports can be separate by `,`.")
	cmd.Flags().StringVarP(&machineName, "name", "n", "", "Name for the instance")
}

func correctArguments(machineVersion string, machineArch string, machineCPU string,
	machineMemory string, machineDisk string, sshPort string, machinePort string) error {

	if machineVersion != "3.16.0" {
		return errors.New("unsupported version. only -v 3.16.0 is currently available")
	}

	if machineArch != "aarch64" && machineArch != "x86_64" {
		return errors.New("unsupported machine architecture. use x86_64 or aarch64")
	}

	int, err := strconv.Atoi(machineCPU)
	if err != nil || int < 0 {
		return errors.New("number of cpus (-c) must be a positive integer")
	}

	int, err = strconv.Atoi(machineMemory)
	if err != nil || int < 250 {
		return errors.New("machine memory (-m) must be a positive integer greater than 250")
	}

	var l, n []rune
	for _, r := range machineDisk {
		switch {
		case r >= 'A' && r <= 'Z':
			l = append(l, r)
		case r >= '0' && r <= '9':
			n = append(n, r)
		}
	}

	int, err = strconv.Atoi(string(n))
	if err != nil || int < 0 {
		return errors.New("disk size (-d) must be a positive integer followed by either K, M, G suffix")
	}

	if !utils.StringSliceContains([]string{"K", "M", "G"}, string(l)) {
		return errors.New("disk size suffix must be K, M, or G")
	}

	int, err = strconv.Atoi(sshPort)
	if err != nil || int < 0 {
		return errors.New("ssh port (-s) must be a positive integer")
	}

	ports := strings.Split(machinePort, ",")

	if machinePort != "" {
		for _, p := range ports {
			int, err = strconv.Atoi(p)
			if err != nil || int < 0 {
				return errors.New("port must be positive integer separated by commas without spaces")
			}
		}
	}

	if machineMount != "" {
		if _, err := os.Stat(machineMount); os.IsNotExist(err) {
			return errors.New("mount directory " + machineMount + " does not exist")
		}
	}

	return nil
}

func launch(cmd *cobra.Command, args []string) {

	err := correctArguments(machineVersion, machineArch, machineCPU, machineMemory, machineDisk, sshPort, machinePort)
	if err != nil {
		log.Fatal("parameter format: " + err.Error())
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if machineName == "" {
		machineName = utils.GenerateRandomAlias()
	}

	vmList, err := host.ListVMNames()
	if err != nil {
		log.Fatal(err)
	}

	exists := utils.StringSliceContains(vmList, machineName)
	if exists {
		log.Fatal("machine " + machineName + " exists")
	}

	if machineMount == "" {
		machineMount = userHomeDir
	}

	machineConfig := qemu.MachineConfig{
		Alias:   machineName,
		Image:   "alpine_" + machineVersion + "-" + machineArch + ".qcow2",
		Arch:    machineArch,
		Version: machineVersion,
		CPU:     machineCPU,
		Memory:  machineMemory,
		Disk:    machineDisk,
		Mount:   machineMount,
		Port:    machinePort,
		SSHPort: sshPort,
	}
	machineConfig.Location = filepath.Join(userHomeDir, ".macpine", machineConfig.Alias)

	err = host.Launch(machineConfig)
	if err != nil {

		os.RemoveAll(machineConfig.Location)
		log.Fatal(err)
	}

	fmt.Println("Launched:", machineName)

}
