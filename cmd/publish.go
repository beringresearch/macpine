package cmd

import (
	"io/ioutil"
	"log"
	"os"
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

	ValidArgsFunction:     host.AutoCompleteVMNames,
	DisableFlagsInUseLine: true,
}

func publish(cmd *cobra.Command, args []string) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("missing VM name")
	}

	vmList, err := host.ListVMNames()
	if err != nil {
		log.Fatal(err)
	}

	exists := utils.StringSliceContains(vmList, args[0])
	if !exists {
		log.Fatal("unknown machine " + args[0])
	}

	machineConfig := qemu.MachineConfig{}

	config, err := ioutil.ReadFile(filepath.Join(userHomeDir, ".macpine", args[0], "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = host.Stop(machineConfig)
	if err != nil {
		log.Fatal(err)
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

	out, err := os.Create(machineConfig.Alias + ".tar.gz")
	if err != nil {
		log.Fatalln("Error writing archive:", err)
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = utils.Compress(files, out)
	if err != nil {
		log.Fatalln("error creating archive:", err)
	}

	err = host.Start(machineConfig)
	if err != nil {
		log.Fatal(err)
	}

}
