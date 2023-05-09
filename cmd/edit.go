package cmd

import (
	"errors"
	"log"
	"os"
	run "os/exec"
	"path/filepath"
	"strings"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// editCmd lists Alpine instances
var editCmd = &cobra.Command{
	Use:     "edit <instance> [<instance>...]",
	Short:   "Edit instance configurations.",
	Run:     edit,
	Aliases: []string{"conf", "config", "configure"},

	ValidArgsFunction:     host.AutoCompleteVMNamesOrTags,
	DisableFlagsInUseLine: true,
}

func edit(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing instance name")
	}

	args, err = host.ExpandTagArguments(args)
	if err != nil {
		log.Fatalln(err)
	}

	vmList := host.ListVMNames()
	for i, vmName := range args {
		if utils.StringSliceContains(args[:i], vmName) {
			continue
		}
		exists := utils.StringSliceContains(vmList, vmName)
		if !exists {
			log.Fatalln("unknown instance " + vmName)
		}
	}

	targetFiles := make([]string, len(args))
	for i, name := range args {
		targetFiles[i] = filepath.Join(userHomeDir, ".macpine", name, "config.yaml")
	}

	oldConfigs := make([]qemu.MachineConfig, len(args))
	for i, vmName := range args {
		oldConfig, err := qemu.GetMachineConfig(vmName)
		if err != nil {
			log.Fatalf("error reading existing configuration for %s\n", vmName)
		}
		oldConfigs[i] = oldConfig
	}

	editor, found := os.LookupEnv("EDITOR")
	if !found || !utils.CommandExists(editor) {
		if !found {
			log.Println("edit: No $EDITOR set.")
		} else {
			log.Println("edit: $EDITOR set but not found in $PATH.")
		}
		if utils.CommandExists("vim") {
			log.Println("defaulting to \"vim\"")
			editor = "vim"
		} else if utils.CommandExists("nano") {
			log.Println("defaulting to \"nano\"")
			editor = "nano"
		} else {
			log.Fatal("no basic editor found in $PATH (tried vim, nano). You can still edit the config manually in ~/.macpine")
		}
	}

	edit := run.Command(editor, targetFiles...)

	edit.Stdin = os.Stdin
	edit.Stdout = os.Stdout
	edit.Stderr = os.Stderr

	err = edit.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = edit.Wait()
	if err != nil {
		log.Fatalf("error while editing: %v\n", err)
	}

	errs := validateConfig(args, targetFiles)
	wasErr := false
	for i, res := range errs {
		if res.Err != nil {
			log.Printf("error in %s configuration: %v\n", res.Name, res.Err)
			log.Printf("reverting %s configuration file\n", res.Name)
			qemu.SaveMachineConfig(oldConfigs[i])
			wasErr = true
		}
	}
	log.Println("configuration(s) saved, restart instance(s) for changes to take effect")
	if wasErr {
		log.Fatalln("error editing instance configuration(s)")
	}
}

func validateConfig(args []string, targetFiles []string) []utils.CmdResult {
	errs := make([]utils.CmdResult, len(args))
	for i := 0; i < len(args); i++ {
		vmName := args[i]
		machineConfig, err := qemu.GetMachineConfig(vmName)

		err = ValidateName(machineConfig.Alias)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}
		image := strings.TrimSuffix(machineConfig.Image, "-"+machineConfig.Arch+".qcow2")
		err = CorrectArguments(image, machineConfig.Arch, machineConfig.CPU,
			machineConfig.Memory, machineConfig.Disk, machineConfig.SSHPort,
			machineConfig.Port)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}
		if loc, err := os.Stat(machineConfig.Location); os.IsNotExist(err) {
			errs[i] = utils.CmdResult{Name: vmName, Err: errors.New("location directory does not exist")}
			continue
		} else if !loc.IsDir() {
			errs[i] = utils.CmdResult{Name: vmName, Err: errors.New("location file is not a directory")}
			continue
		}
	}
	return errs
}
