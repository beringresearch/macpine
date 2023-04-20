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
		log.Fatal(err)
	}

	archive := args[0]

	if strings.HasSuffix(archive, ".age") {
		err = decryptArchive(archive)
		if err != nil {
			log.Fatalln(err)
		}
		archive = strings.TrimSuffix(archive, ".age")
	}

	_, err = utils.CopyFile(archive, filepath.Join(userHomeDir, ".macpine", archive))
	if err != nil {
		log.Fatal(err)
	}

	targetDir := filepath.Join(userHomeDir, ".macpine", strings.Split(archive, ".tar.gz")[0])
	err = utils.Uncompress(filepath.Join(userHomeDir, ".macpine", archive), targetDir)
	if err != nil {
		log.Fatal(err)
	}

	machineConfig := qemu.MachineConfig{}
	config, err := ioutil.ReadFile(filepath.Join(targetDir, "config.yaml"))
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal(err)
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal(err)
	}

	machineConfig.Alias = strings.Split(archive, ".tar.gz")[0]
	machineConfig.Location = targetDir

	updatedConfig, err := yaml.Marshal(&machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(targetDir, "config.yaml"), updatedConfig, 0644)
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal(err)
	}

	if err != nil {
		err = os.Remove(filepath.Join(userHomeDir, ".macpine", archive))
		if err != nil {
			log.Fatal("unable to import: " + err.Error())
		}

		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}

	err = os.Remove(filepath.Join(userHomeDir, ".macpine", archive))
	if err != nil {
		os.RemoveAll(targetDir)
		log.Fatal("unable to import: " + err.Error())
	}
}

func decryptArchive(archive string) error {
	var id age.Identity
	pass, err := utils.PassphrasePromptForDecryption()
	if err != nil {
		return err
	}
	id, err = age.NewScryptIdentity(pass)
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

	err = os.Remove(archive)
	if err != nil {
		log.Printf("error deleting archive after decrypt: %v\n", err)
	}
	return nil
}
