package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const longCompletion = `Generate shell autocompletions. Valid arguments are bash, zsh, and fish.`

var (
	// Todo --noDesc param?
	noDesc = false

	shells = []string{"bash", "zsh", "fish", "powershell"}

	// completionCmd creates completion shell files
	completionCmd = &cobra.Command{
		Use:                   fmt.Sprintf("completion [%s]", strings.Join(shells, "|")),
		Short:                 "Generate shell autocompletions",
		Long:                  longCompletion,
		DisableSuggestions:    false,
		DisableFlagsInUseLine: true,
		ValidArgs:             shells,
		Run:                   completion,
		Hidden:                true,
	}
)

func completion(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatal("missing shell")
	}

	switch args[0] {
	case "bash":
		cmd.Root().GenBashCompletionV2(os.Stdout, !noDesc)
	case "zsh":
		if noDesc {
			cmd.Root().GenZshCompletionNoDesc(os.Stdout)
		} else {
			cmd.Root().GenZshCompletion(os.Stdout)
		}
	case "fish":
		cmd.Root().GenFishCompletion(os.Stdout, !noDesc)
	case "powershell":
		if noDesc {
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		} else {
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	}
}
