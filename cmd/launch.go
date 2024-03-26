package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// launchCmd launches an Alpine instance
var launchCmd = &cobra.Command{
	Use:     "launch",
	Short:   "Create and start an instance.",
	Run:     launch,
	Aliases: []string{"create", "new", "l"},

	ValidArgsFunction: flagsLaunch,
}

var machineArch, imageVersion, machineCPU, machineMemory, machineDisk, machinePort, sshPort, machineName, machineMount string
var vmnet bool

func init() {
	includeLaunchFlags(launchCmd)
}

func includeLaunchFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&imageVersion, "image", "i", "alpine_3.16.0", "Image to be launched.")
	cmd.Flags().StringVarP(&machineArch, "arch", "a", "", "Machine architecture. Defaults to host architecture.")
	cmd.Flags().StringVarP(&machineCPU, "cpu", "c", "2", "Number of CPUs to allocate.")
	cmd.Flags().StringVarP(&machineMemory, "memory", "m", "2048", "Amount of memory (in kB) to allocate.")
	cmd.Flags().StringVarP(&machineDisk, "disk", "d", "10G", "Disk space (in bytes) to allocate. K, M, G suffixes are supported.")
	cmd.Flags().StringVar(&machineMount, "mount", "", "Path to a host directory to be shared with the instance.")
	cmd.Flags().StringVarP(&sshPort, "ssh", "s", "22", "Host port to forward for SSH (required).")
	cmd.Flags().StringVarP(&machinePort, "port", "p", "", "Forward additional host ports. Multiple ports can be separated by `,`.")
	cmd.Flags().StringVarP(&machineName, "name", "n", "", "Instance name for use in `alpine` commands.")
	cmd.Flags().BoolVarP(&vmnet, "shared", "v", false, "Toggle whether to use mac's native vmnet-shared mode.")
}

func CorrectArguments(imageVersion string, machineArch string, machineCPU string,
	machineMemory string, machineDisk string, sshPort string, machinePort string) error {

	if !utils.StringSliceContains([]string{"alpine_3.16.0", "alpine_3.19.1", "alpine_3.16.0_lxd", "debian_11.3.0"}, imageVersion) {
		return errors.New("unsupported image. only -i alpine_3.16.0 | alpine_3.19.1 | debian_11.3.0 are currently available")
	}

	if machineArch != "" {
		if machineArch != "aarch64" && machineArch != "x86_64" {
			return errors.New("unsupported guest architecture. use x86_64 or aarch64")
		}
	}

	int, err := strconv.Atoi(machineCPU)
	if err != nil || int < 0 {
		return errors.New("number of cpus (-c) must be a positive integer")
	}

	int, err = strconv.Atoi(machineMemory)
	if err != nil || int < 256 {
		return errors.New("memory (-m) must be a positive integer greater than 256")
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
		return errors.New("disk size (-d) must be a positive integer optionally followed by K, M, or G")
	}

	if !utils.StringSliceContains([]string{"", "K", "M", "G"}, string(l)) {
		return errors.New("disk size suffix must be K, M, or G")
	}

	int, err = strconv.Atoi(sshPort)
	if err != nil || int < 0 {
		return errors.New("ssh port (-s) must be a positive integer")
	}

	_, err = utils.ParsePort(machinePort)
	if err != nil {
		return err
	}

	if machineMount != "" {
		if dir, err := os.Stat(machineMount); os.IsNotExist(err) {
			return errors.New("mount target " + machineMount + " does not exist")
		} else if !dir.IsDir() {
			return errors.New("mount target " + machineMount + " is not a directory")
		}
	}

	return nil
}

func launch(cmd *cobra.Command, args []string) {

	err := CorrectArguments(imageVersion, machineArch, machineCPU, machineMemory, machineDisk, sshPort, machinePort)
	if err != nil {
		log.Fatalln(err.Error())
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	if machineArch == "" {
		arch := runtime.GOARCH

		switch arch {
		case "arm64":
			machineArch = "aarch64"
		case "amd64":
			machineArch = "x86_64"
		default:
			log.Fatal("unsupported host architecture: " + arch)
		}
	}

	vmList := host.ListVMNames()

	if machineName == "" {
		machineName = utils.GenerateRandomAlias()
		for utils.StringSliceContains(vmList, machineName) { // if exists, re-randomize
			machineName = utils.GenerateRandomAlias()
		}
	} else if utils.StringSliceContains(vmList, machineName) {
		log.Fatal("instance with name \"" + machineName + "\" already exists")
	}

	macAddress, err := utils.GenerateMACAddress()
	if err != nil {
		log.Fatal(err)
	}

	machineIP := "localhost"

	machineConfig := qemu.MachineConfig{
		Alias:       machineName,
		Image:       imageVersion + "-" + machineArch + ".qcow2",
		Arch:        machineArch,
		CPU:         machineCPU,
		Memory:      machineMemory,
		Disk:        machineDisk,
		Mount:       machineMount,
		MachineIP:   machineIP,
		Port:        machinePort,
		SSHPort:     sshPort,
		MACAddress:  macAddress,
		VMNet:       vmnet,
		SSHUser:     "root",
		SSHPassword: "raw::root",
		Tags:        []string{},
	}
	machineConfig.Location = filepath.Join(userHomeDir, ".macpine", machineConfig.Alias)

	err = host.Launch(machineConfig)
	if err != nil {
		os.RemoveAll(machineConfig.Location)
		pid, _ := machineConfig.GetInstancePID()
		p, _ := os.FindProcess(pid)
		p.Signal(syscall.SIGKILL)
		log.Fatal(err)
	}

	fmt.Println("launched: " + machineName)
}

func flagsLaunch(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
