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
	SSHPort  string `yaml:"sshport"`
	Location string `yaml:"location"`
}

//Status returns VM status

func (c *MachineConfig) Status() (string, int) {
	status := ""
	var pid int

	pidFile := filepath.Join(c.Location, "alpine.pid")

	if _, err := os.Stat(pidFile); err == nil {
		status = "Running"
		vmPID, err := ioutil.ReadFile(pidFile)

		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}

		process := string(vmPID)
		process = strings.TrimSuffix(process, "\n")
		pid, _ = strconv.Atoi(process)
	} else {
		status = "Stopped"
	}
	return status, pid
}

//Stop stops an Alpine VM
func (c *MachineConfig) Stop() error {
	pidFile := filepath.Join(c.Location, "alpine.pid")
	if _, err := os.Stat(pidFile); err == nil {

		_, pid := c.Status()

		if pid > 0 {

			err = syscall.Kill(pid, 15)
			if err != nil {
				return err
			}
		} else {
			return errors.New("unable to SIGTERM 15: incorrect PID")
		}
	}

	return nil
}

// Start starts up an Alpine VM
func (c *MachineConfig) Start() error {

	exposedPorts := "user,id=net0,hostfwd=tcp::" + c.SSHPort + "-:22"

	if c.Port != "" {
		s := strings.Split(c.Port, ",")
		for _, p := range s {
			exposedPorts += ",hostfwd=tcp::" + p + "-:" + p
		}
	}

	fmt.Println(exposedPorts)

	cmd := exec.Command("qemu-system-x86_64",
		"-m", c.Memory,
		// use tcg accelerator with multi threading with 512MB translation block size
		// https://qemu-project.gitlab.io/qemu/devel/multi-thread-tcg.html?highlight=tcg
		// https://qemu-project.gitlab.io/qemu/system/invocation.html?highlight=tcg%20opts
		// this will make sure each vCPU will be backed by 1 host user thread.
		"-accel", "tcg,thread=multi,tb-size=512",
		//disable CPU S3 state.
		"-global", "ICH9-LPC.disable_s3=1",
		"-smp", c.CPU+",sockets=1,cores="+c.CPU+",threads=1",
		"-hda", filepath.Join(c.Location, c.Image),
		"-device", "e1000,netdev=net0",
		"-netdev", exposedPorts,
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
		return err
	}

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

	err = c.ResizeQemuDiskImage()
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

	fmt.Println(string(config))

	return nil
}

//ResizeQemuDiskImage resizes a qcow2 disk image
func (c *MachineConfig) ResizeQemuDiskImage() error {
	if !utils.CommandExists("qemu-img") {
		return errors.New("qemu-img is not available on $PATH. ensure qemu is installed")
	}

	cmd := exec.Command("qemu-img",
		"resize",
		filepath.Join(c.Location, c.Image),
		"+"+c.Disk)

	err := cmd.Run()

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
