package cmd

import (
	"io/ioutil"
	"log"
	"os"
	osexec "os/exec"
	"path/filepath"
	"time"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// publishCmd stops an Alpine instance
var publishCmd = &cobra.Command{
	Use:   "publish NAME",
	Short: "Publish an instance.",
	Run:   publish,

	ValidArgsFunction: flagsPublish,
}

var ageRecipient, ageIdent string
var agePw bool

func init() {
	includePublishFlags(publishCmd)
}

func includePublishFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&agePw, "password", "p", false, "Encrypt published VM with interactive passphrase prompt (symmetric).")
	cmd.Flags().StringVarP(&ageIdent, "private-key", "s", "", "Encrypt published VM with ssh/age secret key file (symmetric).")
	cmd.Flags().StringVarP(&ageRecipient, "public-key", "k", "", "Encrypt published VM with ssh/age public key file (asymmetric).")
	cmd.MarkFlagsMutuallyExclusive("password", "private-key", "public-key")
}

func publish(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	if len(args) == 0 {
		log.Fatalln("missing VM name")
	}

	vmList := host.ListVMNames()

	exists := utils.StringSliceContains(vmList, args[0])
	if !exists {
		log.Fatalln("unknown machine " + args[0])
	}

	machineConfig := qemu.MachineConfig{}

	config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", args[0], "config.yaml"))
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		log.Fatalln(err)
	}

	if status, _ := host.Status(machineConfig); status != "Stopped" {
		err = host.Stop(machineConfig)
		if err != nil {
			log.Fatalln(err)
		}
	}

	time.Sleep(time.Second)

	fileInfo, err := ioutil.ReadDir(machineConfig.Location)
	if err != nil {
		log.Fatal(err)
	}

	files := []string{}
	for _, f := range fileInfo {
		files = append(files, filepath.Join(machineConfig.Location, f.Name()))
	}

	gzName := machineConfig.Alias + ".tar.gz"
	out, err := os.Create(gzName)
	if err != nil {
		log.Fatalln("Error writing archive:", err)
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = utils.Compress(files, out)
	if err != nil {
		log.Fatalln("error creating archive:", err)
	}

	if agePw || ageIdent != "" || ageRecipient != "" {
		if !utils.CommandExists("age") {
			log.Fatalln("encryption requested but `age` not installed or not in path")
		}
		var ageCmd *osexec.Cmd
		if agePw {
			ageCmd = osexec.Command("age", "-e", "-p", "-o", gzName+".age", gzName)
		} else if ageIdent != "" {
			ageCmd = osexec.Command("age", "-e", "-i", ageIdent, "-o", gzName+".age", gzName)
		} else if ageRecipient != "" {
			ageCmd = osexec.Command("age", "-e", "-R", ageRecipient, "-o", gzName+".age", gzName)
		}
		if err := ageCmd.Run(); err != nil {
			log.Fatalln(err)
		}
		if err := os.Remove(gzName); err != nil {
			log.Fatalln(err)
		}
	}
}

func flagsPublish(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	flags := []string{"-p", "--password", "-s", "--private-key", "-k", "--public-key"}
	vmnames, cobradefault := host.AutoCompleteVMNames(cmd, args, toComplete)
	flags = append(flags, vmnames...)
	return flags, cobradefault
}
