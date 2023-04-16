package cmd

import (
	"log"
	"os"
	run "os/exec"
	"path/filepath"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// editCmd lists Alpine instances
var editCmd = &cobra.Command{
	Use:   "edit <instance> [<instance>...]",
	Short: "Edit instance configuration.",
	Run:   edit,

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
	} else {
		log.Println("configuration(s) saved, restart instance(s) for changes to take effect")
	}
}
