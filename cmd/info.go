package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/beringresearch/macpine/host"
)

// infoCmd displays macpine machine info
var infoCmd = &cobra.Command{
	Use:   "info NAME",
	Short: "Display information about an instance.",
	Run:   macpineInfo,

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func macpineInfo(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatal("missing VM name")
	}

	for _, vmName := range args {
		info, err := host.Info(vmName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(info)
	}

}
