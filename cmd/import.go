package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	qemu "github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// importCmd iports an Alpine VM from file
var importCmd = &cobra.Command{
	Use:   "import <archive>",
	Short: "Imports an instance archived with `alpine publish`.",
	Run:   importMachine,

	DisableFlagsInUseLine: true,
}

func importMachine(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		log.Fatal("missing archive filename")
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("unable to import: " + err.Error())
	}

	archive := args[0]

	if strings.HasSuffix(archive, ".age") {
		err = decryptArchive(archive)
		if err != nil {
			log.Fatal("unable to import: " + err.Error())
		}
		archive = strings.TrimSuffix(archive, ".age")
	}
	importName := strings.TrimSuffix(archive, ".tar.gz")
	tempArchive := filepath.Join(userHomeDir, ".macpine", archive)

	_, err = utils.CopyFile(archive, tempArchive)
	if err != nil {
		os.RemoveAll(tempArchive)
		log.Fatal("unable to import: " + err.Error())
	}
	defer os.RemoveAll(tempArchive)

	targetDir := strings.TrimSuffix(tempArchive, ".tar.gz")
	exists, err := utils.DirExists(targetDir)
	if err != nil {
		log.Fatal("unable to import: " + err.Error())
	}
	if exists {
		log.Fatalf("unable to import: instance %s already exists\n", importName)
	}
	err = utils.Uncompress(tempArchive, targetDir)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}

	machineConfig := qemu.MachineConfig{}
	config, err := ioutil.ReadFile(filepath.Join(targetDir, "config.yaml"))
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}

	machineConfig.Alias = importName
	machineConfig.Location = targetDir

	updatedConfig, err := yaml.Marshal(&machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}

	err = ioutil.WriteFile(filepath.Join(targetDir, "config.yaml"), updatedConfig, 0644)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}
}

func decryptArchive(archive string) error {
	pass, err := utils.PassphrasePromptForDecryption()
	if err != nil {
		return err
	}
	id, err := age.NewScryptIdentity(pass)
	if err != nil {
		return fmt.Errorf("error generating key for passphrase: %v\n", err)
	}
	src, err := os.Open(archive)
	if err != nil {
		return fmt.Errorf("error reading encrypted archive: %v\n", err)
	}
	dst, err := os.Create(strings.TrimSuffix(archive, ".age"))
	if err != nil {
		return fmt.Errorf("error creating decrypted archive: %v\n", err)
	}
	defer dst.Close()

	dec, err := age.Decrypt(src, id)
	if err != nil {
		return fmt.Errorf("error decrypting archive: %v\n", err)
	}
	data, err := io.ReadAll(dec)
	if err != nil {
		return fmt.Errorf("error reading decrypted archive: %v\n", err)
	}

	_, err = dst.Write(data)
	if err != nil {
		return fmt.Errorf("error writing decrypted archive: %v\n", archive)
	}

	return nil
}
