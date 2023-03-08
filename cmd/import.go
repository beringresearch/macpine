package cmd

import (
	"io/ioutil"
	"log"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beringresearch/macpine/host"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// importCmd iports an Alpine VM from file
var importCmd = &cobra.Command{
	Use:   "import FILE",
	Short: "Imports an instance.",
	Run:   importMachine,
}

var agePrivate string

func init() {
	includeImportFlags(importCmd)
}

func includeImportFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&agePrivate, "private-key", "s", "", "Private key file to decrypt the imported archive.")
}

func importMachine(cmd *cobra.Command, args []string) {

	if len(args) == 0 || (len(args) == 1 && args[0] == "") {
		log.Fatalln("missing archive file name")
	}
	archive := args[0]
	if _, err := os.Stat(archive); err != nil {
		if os.IsNotExist(err) {
			log.Fatalln("no archive file found")
		} else {
			log.Fatalln("error accessing archive file: " + err.Error())
		}
	}
	if (strings.HasSuffix(archive, ".age") || agePrivate != "") && !utils.CommandExists("age") {
		log.Fatalln("decryption requested but `age` not installed or not in path")
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	_, err = utils.CopyFile(archive, filepath.Join(userHomeDir, ".macpine", filepath.Base(archive)))
	if err != nil {
		log.Fatalln(err)
	}
	archive = filepath.Join(userHomeDir, ".macpine", filepath.Base(archive))
	defer os.Remove(archive) // delete the copied archive when done

	if strings.HasSuffix(archive, ".age") {
		decryptName := strings.TrimSuffix(archive, ".age")
		// archive is encrypted, decrypt with password or private key
		var ageCmd *osexec.Cmd
		if agePrivate == "" { // no key specified, assume a password-encrypted archive
			ageCmd = osexec.Command("age", "-d", "-p", "-o", decryptName, archive)
		} else {
			ageCmd = osexec.Command("age", "-d", "-i", agePrivate, "-o", decryptName, archive)
		}
		if err := ageCmd.Run(); err != nil {
			log.Fatalln("failed to decrypt archive: " + err.Error())
		}
		defer os.Remove(decryptName) // delete the decrypted archive copy when done
		archive = decryptName
	}

	name := strings.TrimSuffix(filepath.Base(archive), ".tar.gz")
	if utils.StringSliceContains(host.ListVMNames(), name) {
		var nameExtra int
		for nameExtra = 1; utils.StringSliceContains(host.ListVMNames(), name+"-"+strconv.Itoa(nameExtra)); nameExtra += 1 {
			if nameExtra >= 1024 {
				log.Fatalf("Too many VMs with name prefix %s found. Please rename the archive to retry.\n", name)
			}
		}
		name = name + "-" + strconv.Itoa(nameExtra)
	}
	log.Printf("importing %s as %s\n", filepath.Base(archive), name)

	targetDir := filepath.Join(userHomeDir, ".macpine", name)
	err = utils.Uncompress(archive, targetDir)
	if err != nil {
		log.Fatalln("failed to expand archive: " + err.Error())
	}

	machineConfig := qemu.MachineConfig{}
	config, err := ioutil.ReadFile(filepath.Join(targetDir, "config.yaml"))
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatalln("failed to parse archive: " + err.Error())
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatalln("failed to parse archive: " + err.Error())
	}

	machineConfig.Alias = name
	machineConfig.Location = targetDir

	updatedConfig, err := yaml.Marshal(&machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatalln("failed to update config: " + err.Error())
	}

	err = ioutil.WriteFile(filepath.Join(targetDir, "config.yaml"), updatedConfig, 0644)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatalln("failed to update config: " + err.Error())
	}

	log.Println("successfully imported " + name)
}
