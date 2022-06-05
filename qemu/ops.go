package qemu

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/beringresearch/macpine/utils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
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

//ExecShell starts an interactive shell terminal in VM
func (c *MachineConfig) ExecShell() error {

	host := "localhost:" + c.SSHPort
	user := "root"
	pwd := "root"

	var err error

	conf := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
	}

	var conn *ssh.Client

	conn, err = ssh.Dial("tcp", host, conf)
	if err != nil {
		return err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	in, _ := session.StdinPipe()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	width, height, err := term.GetSize(0)
	if err != nil {
		return err
	}

	if err := session.RequestPty("xterm", height, width, modes); err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		str, _ := reader.ReadString('\n')
		fmt.Fprint(in, str)
	}
}

//Exec executes command inside VM
func (c *MachineConfig) Exec(cmd string) error {

	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password("root"),
		},
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort("localhost", c.SSHPort), config)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	stdin := new(bytes.Buffer)

	session.Stdout = stdout
	session.Stderr = stderr
	session.Stdin = stdin

	if err := session.Run(cmd); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	fmt.Println(stdout.String())

	return nil
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

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	exposedPorts := "user,id=net0,hostfwd=tcp::" + c.SSHPort + "-:22"

	if c.Port != "" {
		s := strings.Split(c.Port, ",")
		for _, p := range s {
			exposedPorts += ",hostfwd=tcp::" + p + "-:" + p
		}
	}

	qemuCmd := "qemu-system-" + c.Arch

	var qemuArgs []string

	accelAarch64 := "hvf"
	cpuAarch64 := "cortex-a72"

	if runtime.GOOS == "linux" {
		accelAarch64 = "tcg"
		cpuAarch64 = "cortex-a57"
	}
	aarch64Args := []string{
		//"-cpu", "host",
		"-accel", accelAarch64,
		"-cpu", cpuAarch64,
		"-M", "virt,highmem=off",
		"-bios", filepath.Join(c.Location, "qemu_efi.fd")}

	accelx86 := "tcg"
	if runtime.GOOS == "darwin" {
		accelx86 += ",thread=multi,tb-size=512"
	}
	x86Args := []string{"-accel", accelx86}

	commonArgs := []string{"-m", c.Memory, "-global", "ICH9-LPC.disable_s3=1",
		"-smp", c.CPU + ",sockets=1,cores=" + c.CPU + ",threads=1",
		"-hda", filepath.Join(c.Location, c.Image),
		"-nographic",
		"-device", "e1000,netdev=net0",
		"-netdev", exposedPorts,
		"-pidfile", filepath.Join(c.Location, "alpine.pid"),
		"-chardev", "socket,id=char-serial,path=" + filepath.Join(c.Location,
			"alpine.sock") + ",server=on,wait=off,logfile=" + filepath.Join(c.Location, "alpine.log"),
		"-serial", "chardev:char-serial",
		"-chardev", "socket,id=char-qmp,path=" + filepath.Join(c.Location, "alpine.qmp") + ",server=on,wait=off",
		"-qmp", "chardev:char-qmp",
		"-parallel", "none",
		"-virtfs", "local,path=" + userHomeDir + ",security_model=none,mount_tag=Home",
		"-name", "alpine"}

	if c.Arch == "aarch64" {
		qemuArgs = append(aarch64Args, commonArgs...)
	}
	if c.Arch == "x86_64" {
		qemuArgs = append(x86Args, commonArgs...)
	}

	cmd := exec.Command(qemuCmd, qemuArgs...,
	)

	cmd.Stdout = os.Stdout

	// Uncomment to debug qemu messages
	//cmd.Stderr = os.Stderr

	log.Printf("Booting...")
	err = cmd.Start()
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	err = utils.Retry(10, 2*time.Second, func() error {
		err := c.Exec("mkdir -p /root/mnt/; mount -t 9p -o trans=virtio Home /root/mnt/ -oversion=9p2000.L")
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return errors.New("unable to mount: " + err.Error())
	}

	log.Printf("Mounted file system on: /root/mnt/")

	status, _ := c.Status()

	if status == "Stopped" {
		return errors.New("unable to start vm")
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

	if c.Arch == "aarch64" {
		if _, err := os.Stat(filepath.Join(cacheDir, "qemu_efi.fd")); errors.Is(err, os.ErrNotExist) {
			err = utils.DownloadFile(filepath.Join(cacheDir, "qemu_efi.fd"),
				"https://github.com/beringresearch/macpine/releases/download/v.01/qemu_efi.fd")
			if err != nil {
				return errors.New("unable to download bios :" + err.Error())
			}
		}
	}

	targetDir := filepath.Join(userHomeDir, ".macpine", c.Alias)
	err = os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return err
	}

	_, err = utils.CopyFile(filepath.Join(cacheDir, imageName), filepath.Join(targetDir, imageName))
	if err != nil {
		os.RemoveAll(targetDir)
		return err
	}

	if c.Arch == "aarch64" {
		_, err = utils.CopyFile(filepath.Join(cacheDir, "qemu_efi.fd"), filepath.Join(targetDir, "qemu_efi.fd"))
		if err != nil {
			os.RemoveAll(targetDir)
			return err
		}
	}

	err = c.ResizeQemuDiskImage()
	if err != nil {
		os.RemoveAll(targetDir)
		return errors.New("unable to resize disk: " + err.Error())
	}

	config, err := yaml.Marshal(&c)

	if err != nil {
		os.RemoveAll(targetDir)
		return err
	}

	err = ioutil.WriteFile(filepath.Join(c.Location, "config.yaml"), config, 0644)
	if err != nil {
		os.RemoveAll(targetDir)
		return err
	}

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
