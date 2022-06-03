package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// importCmd iports an Alpine VM from file
var importCmd = &cobra.Command{
	Use:   "import NAME",
	Short: "Imports Alpine VM instances.",
	Run:   importMachine,
}

func importMachine(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		log.Fatal("missing VN name")
		return
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	_, err = utils.CopyFile(args[0], filepath.Join(userHomeDir, ".macpine", args[0]))
	if err != nil {
		log.Fatal(err)
	}

	targetDir := filepath.Join(userHomeDir, ".macpine", strings.Split(args[0], ".tar.gz")[0])
	err = utils.Uncompress(filepath.Join(userHomeDir, ".macpine", args[0]), targetDir)

	if err != nil {
		err = os.Remove(filepath.Join(userHomeDir, ".macpine", args[0]))
		if err != nil {
			log.Fatal("unable to import: " + err.Error())
		}

		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}

	err = os.Remove(filepath.Join(userHomeDir, ".macpine", args[0]))
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}
}
