package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"filippo.io/age"
	"github.com/beringresearch/macpine/host"
	"github.com/beringresearch/macpine/qemu"
	"github.com/beringresearch/macpine/utils"
	"github.com/spf13/cobra"
)

// publishCmd stops an Alpine instance
var publishCmd = &cobra.Command{
	Use:     "publish <instance> [<instance>...]",
	Short:   "Publish instances.",
	Run:     publish,
	Aliases: []string{"pub", "archive"},

	ValidArgsFunction: host.AutoCompleteVMNamesOrTags,
}

var encrypt bool

func init() {
	includePublishFlags(publishCmd)
}

func includePublishFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&encrypt, "encrypt", "e", false, "Encrypt published archive (prompts for passphrase).")
}

func publish(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		log.Fatal("missing instance name")
	}

	args, err := host.ExpandTagArguments(args)
	if err != nil {
		log.Fatal("unable to publish: " + err.Error())
	}

	vmList := host.ListVMNames()
	errs := make([]utils.CmdResult, len(args))
	for i, vmName := range args {
		if utils.StringSliceContains(args[:i], vmName) {
			continue
		}
		exists := utils.StringSliceContains(vmList, vmName)
		if !exists {
			errs[i] = utils.CmdResult{Name: vmName, Err: errors.New("unknown instance " + vmName)}
			continue
		}

		machineConfig, err := qemu.GetMachineConfig(vmName)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		vmStatus, _ := host.Status(machineConfig)
		if vmStatus == "Running" {
			err = machineConfig.Exec("sync", true)
			if err != nil {
				errs[i] = utils.CmdResult{
					Name: vmName, Err: errors.New("error synchonizing filesystem before publish, stop instance and retry")}
				continue
			}
			err = host.Pause(machineConfig)
			if err != nil {
				errs[i] = utils.CmdResult{
					Name: vmName, Err: errors.New("error pausing instance before publish, stop instance and retry")}
				continue
			}
			time.Sleep(time.Second)
		}

		fileInfo, err := os.ReadDir(machineConfig.Location)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		files := []string{}
		for _, f := range fileInfo {
			if !utils.StringSliceContains([]string{"alpine.qmp", "alpine.sock", "alpine.pid"}, f.Name()) {
				files = append(files, filepath.Join(machineConfig.Location, f.Name()))
			}
		}

		out, err := os.Create(machineConfig.Alias + ".tar.gz")
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}
		defer out.Close()

		// Create the archive and write the output to the "out" Writer
		ext := ""
		if encrypt {
			ext = ".age"
		}
		log.Printf("creating archive %s...\n", machineConfig.Alias+".tar.gz"+ext)

		err = utils.Compress(files, out)
		if err != nil {
			errs[i] = utils.CmdResult{Name: vmName, Err: err}
			continue
		}

		if encrypt {
			err = encryptArchive(&machineConfig)
			if err != nil {
				errs[i] = utils.CmdResult{Name: vmName, Err: err}
				continue
			}
		}

		if vmStatus == "Running" {
			err = host.Resume(machineConfig)
			if err != nil {
				errs[i] = utils.CmdResult{Name: vmName, Err: err}
				continue
			}
		}
	}
	wasErr := false
	for _, res := range errs {
		if res.Err != nil {
			log.Printf("unable to publish %s: %v\n", res.Name, res.Err)
			wasErr = true
		}
	}
	if wasErr {
		log.Fatalln("error publishing instance(s)")
	}
}

func encryptArchive(machineConfig *qemu.MachineConfig) error {
	pass, err := utils.PassphrasePromptForEncryption()
	if err != nil {
		return err
	}
	rs, err := age.NewScryptRecipient(pass)
	if err != nil {
		return fmt.Errorf("error generating key for passphrase: %v\n", err)
	}
	dst, err := os.Create(machineConfig.Alias + ".tar.gz.age")
	if err != nil {
		return fmt.Errorf("error creating encrypted archive: %v\n", err)
	}
	defer dst.Close()

	enc, err := age.Encrypt(dst, rs)
	if err != nil {
		return fmt.Errorf("error encrypting archive: %v\n", err)
	}
	defer enc.Close()

	archive, err := os.ReadFile(machineConfig.Alias + ".tar.gz")
	if err != nil {
		return fmt.Errorf("error reading archive to encrypt: %v\n", archive)
	}
	_, err = enc.Write(archive)
	if err != nil {
		return fmt.Errorf("error writing archive: %v\n", err)
	}

	err = os.Remove(machineConfig.Alias + ".tar.gz")
	if err != nil {
		return fmt.Errorf("error cleaning up after enryption: %v\n", err)
	}
	return nil
}
