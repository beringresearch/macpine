package cmd

import (
	"log"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
)

var remove bool

func init() {
	includeTagFlag(tagCmd)
}

func includeTagFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&remove, "remove", "r", false, "Remove tag(s) rather than add them.")
}

var tagCmd = &cobra.Command{
	Use:   "tag [-r] <instance> <tag1> [<tag2>...]",
	Short: "Add or remove tags from an instance.",
	Run:   macpineTag,

	ValidArgsFunction: host.AutoCompleteVMNames,
}

func validateTags(tags []string) {
	format := regexp.MustCompile("^[a-zA-Z0-9_\\-]*$")
	for _, tag := range tags {
		if !format.MatchString(tag) {
			log.Fatalf("[%s] contains invalid characters (alphanumeric, _, and -)", tag)
		}
	}
}

func macpineTag(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatal("missing instance name")
	}
	vmName := args[0]

	tags := args[1:]
	validateTags(tags)

	machineConfig, err := qemu.GetMachineConfig(vmName)
	if err != nil {
		log.Fatalln(err)
	}

	for _, tag := range tags {
		i, found := find(machineConfig.Tags, tag)
		if remove && found {
			machineConfig.Tags = append(machineConfig.Tags[:i], machineConfig.Tags[i+1:]...)
		} else if !remove && !found {
			machineConfig.Tags = append(machineConfig.Tags[:i], append([]string{tag}, machineConfig.Tags[i:]...)...)
		}
	}

	err = qemu.SaveMachineConfig(machineConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s tags: "+strings.Join(machineConfig.Tags[:], ", "), machineConfig.Alias)
}

func find(s []string, t string) (int, bool) {
	for i, e := range s {
		if e == t {
			return i, true
		} else if e > t {
			return i, false
		}
	}
	return len(s), false
}
