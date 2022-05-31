package qemu

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/beringresearch/macpine/utils"
	"gopkg.in/yaml.v2"
)

type MachineConfig struct {
	Alias    string `yaml:"alias"`
	Image    string `yaml:"image"`
	Arch     string `yaml:"arch"`
	Version  string `yaml:"version"`
	CPU      string `yaml:"cpu"`
	Memory   string `yaml:"memory"`
	Disk     string `yaml:"disk"`
	Port     string `yaml:"port"`
	Location string `yaml:"location"`
}

//Stop stops an Alpine VM
func (c *MachineConfig) Stop() error {
	pid, err := ioutil.ReadFile(filepath.Join(c.Location, "alpine.pid"))

	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}

	process := string(pid)
	process = strings.TrimSuffix(process, "\n")
	vmPID, _ := strconv.Atoi(process)

	fmt.Println(vmPID)

	err = syscall.Kill(vmPID, 15)
	if err != nil {
		return err
	}

	return nil
}

// Start starts up an Alpine VM
func (c *MachineConfig) Start() error {

	cmd := exec.Command("qemu-system-x86_64",
		"-m", c.Memory,
		"-smp", c.CPU,
		"-hda", filepath.Join(c.Location, c.Image),
		"-device", "e1000,netdev=net0",
		"-netdev", "user,id=net0,hostfwd=tcp::"+c.Port+"-:22",
		"-pidfile", filepath.Join(c.Location, "alpine.pid"),
		"-chardev", "socket,id=char-serial,path="+filepath.Join(c.Location,
			"alpine.sock")+",server=on,wait=off,logfile="+filepath.Join(c.Location, "alpine.log"),
		"-serial", "chardev:char-serial",
		"-chardev", "socket,id=char-qmp,path="+filepath.Join(c.Location, "alpine.qmp")+",server=on,wait=off",
		"-qmp", "chardev:char-qmp",
		"-parallel", "none",
		"-name", "alpine",
	)

	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Just ran subprocess %d, exiting\n", cmd.Process.Pid)

	return nil
}

//Launch macpine downloads a fresh image and creates a VM directory
func (c *MachineConfig) Launch() error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(userHomeDir, ".macpine", "cache")
	err = os.MkdirAll(cacheDir, os.ModePerm)
	if err != nil {
		return err
	}

	imageName, alpineURL := utils.GetAlpineURL(c.Version, c.Arch)

	if _, err := os.Stat(filepath.Join(cacheDir, imageName)); errors.Is(err, os.ErrNotExist) {
		err = utils.DownloadFile(filepath.Join(cacheDir, imageName), alpineURL)
		if err != nil {
			return errors.New("unable to download Alpine " + c.Version + " for " + c.Arch + ": " + err.Error())
		}
	}

	targetDir := filepath.Join(userHomeDir, ".macpine", c.Alias)
	err = os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return err
	}

	_, err = utils.CopyFile(filepath.Join(cacheDir, imageName), filepath.Join(targetDir, imageName))
	if err != nil {
		return err
	}

	config, err := yaml.Marshal(&c)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(c.Location, "config.yaml"), config, 0644)
	if err != nil {

		return err
	}

	return nil
}

//CreateQemuDiskImage creates a qcow2 disk image
func (c *MachineConfig) CreateQemuDiskImage(imageName string) error {

	if !utils.CommandExists("qemu-img") {
		return errors.New("qemu-img is not available on $PATH. ensure qemu is installed")
	}
	cmd := exec.Command("qemu-img",
		"create", "-f", "qcow2",
		"-o", "compression_type=zlib",
		filepath.Join(c.Location, imageName),
		c.Disk)

	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}
