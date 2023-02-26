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
	Use:   "edit NAME",
	Short: "Edit instance configuration.",
	Run:   edit,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func edit(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing VM name")
	}

	vmList, err := host.ListVMNames()
	if err != nil {
		log.Fatal(err)
	}

	exists := utils.StringSliceContains(vmList, args[0])
	if !exists {
		log.Fatal("unknown machine " + args[0])
	}

	targetFile := filepath.Join(userHomeDir, ".macpine", args[0], "config.yaml")

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
         log.Fatal("no basic editor found in $PATH (tried vim, nano). You can still edit the config manually at " + targetFile)
      }
   }

   edit := run.Command(editor, targetFile)

   edit.Stdin = os.Stdin
   edit.Stdout = os.Stdout
   edit.Stderr = os.Stderr

   err = edit.Start()
   if err != nil {
      log.Fatal(err)
   }

   err = edit.Wait()
   if err != nil {
      log.Fatalf("error while editing. Error: %v\n", err)
   } else {
      log.Printf("configuration saved. restart " + args[0] + " for changes to take effect.")
   }
}
