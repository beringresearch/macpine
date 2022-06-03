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

	err = utils.Uncompress(filepath.Join(userHomeDir, ".macpine", args[0]),
		filepath.Join(userHomeDir, ".macpine", strings.Split(args[0], ".tar.gz")[0]))

	if err != nil {
		log.Println(err)
	}

	err = os.Remove(filepath.Join(userHomeDir, ".macpine", args[0]))
	if err != nil {
		log.Fatal(err)
	}
}
