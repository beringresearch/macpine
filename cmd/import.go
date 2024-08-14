package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// importCmd iports an Alpine VM from file
var importCmd = &cobra.Command{
	Use:     "import <archive>",
	Short:   "Imports an instance archived with `alpine publish`. Can be a local import or a URL.",
	Run:     importMachine,
	Aliases: []string{"unarchive"},

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

	macpineHomeDir := filepath.Join(userHomeDir, ".macpine")
	if _, err := os.Stat(macpineHomeDir); os.IsNotExist(err) {
		err := os.Mkdir(macpineHomeDir, 0700)
		if err != nil {
			log.Fatal("$HOME/.macpine directory does not exist and unable to create it: " + err.Error())
		}
	}

	archive := args[0]

	if strings.HasPrefix(archive, "http") {
		//cachePath := filepath.Join(userHomeDir, ".macpine", "cache")
		archiveString := path.Base(archive)

		// Handle a specific case for a Dropbox URL
		archiveString = strings.Split(archiveString, "?rlkey")[0]

		archiveFile, err := os.Create(filepath.Join("/tmp/", archiveString))

		if err != nil {
			log.Fatal("unable to download macpine image: " + err.Error())
		}

		err = utils.DownloadFile(archiveFile.Name(), archive)
		if err != nil {
			log.Fatal("unable to download macpine image: " + err.Error())
		}

		archive = archiveFile.Name()

	}

	if strings.HasSuffix(archive, ".age") {
		err = decryptArchive(archive)
		if err != nil {
			log.Fatal("unable to import: " + err.Error())
		}
		archive = strings.TrimSuffix(archive, ".age")
		defer os.RemoveAll(archive)
	}

	if !strings.HasSuffix(archive, ".tar.gz") {
		log.Fatal("unable to import: instance must be .age or .tar.gz file")
	}

	importName := strings.TrimSuffix(filepath.Base(archive), ".tar.gz")
	tempArchive := filepath.Join(userHomeDir, ".macpine", filepath.Base(archive))

	targetDir := strings.TrimSuffix(tempArchive, ".tar.gz")

	exists, err := utils.DirExists(targetDir)
	if err != nil {
		os.RemoveAll(tempArchive)
		log.Fatal("unable to import: " + err.Error())
	}
	if exists {
		os.RemoveAll(tempArchive)
		log.Fatalf("unable to import: instance %s already exists\n", importName)
	}

	_, err = utils.CopyFile(archive, tempArchive)
	if err != nil {
		os.RemoveAll(tempArchive)
		log.Fatal("unable to import: " + err.Error())
	}
	defer os.RemoveAll(tempArchive)

	err = utils.Uncompress(tempArchive, targetDir)
	if err != nil {
		os.RemoveAll(targetDir)
		os.RemoveAll(tempArchive)
		log.Fatal("unable to import: " + err.Error())
	}

	machineConfig, err := qemu.GetMachineConfig(importName)
	if err != nil {
		os.RemoveAll(targetDir)
		os.RemoveAll(tempArchive)
		log.Fatal("unable to import: " + err.Error())
	}

	machineConfig.Alias = importName
	machineConfig.Location = targetDir
	machineConfig.MachineIP = "localhost"

	err = qemu.SaveMachineConfig(machineConfig)
	if err != nil {
		os.RemoveAll(targetDir)
		os.RemoveAll(tempArchive)
		log.Fatal("unable to import: " + err.Error())
	}

	err = machineConfig.DecompressQemuDiskImage()
	if err != nil {
		os.RemoveAll(targetDir)
		os.RemoveAll(tempArchive)
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
