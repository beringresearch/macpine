package qemu

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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
	Alias       string `yaml:"alias"`
	Image       string `yaml:"image"`
	Arch        string `yaml:"arch"`
	CPU         string `yaml:"cpu"`
	Memory      string `yaml:"memory"`
	Disk        string `yaml:"disk"`
	Mount       string `yaml:"mount"`
	Port        string `yaml:"port"`
	SSHPort     string `yaml:"sshport"`
	SSHUser     string `yaml:"sshuser"`
	SSHPassword string `yaml:"sshpassword"`
	MACAddress  string `yaml:"macaddress"`
	Location    string `yaml:"location"`
}

// Exec starts an interactive shell terminal in VM
func (c *MachineConfig) Exec(cmd string) error {

	host := "localhost:" + c.SSHPort
	user := c.SSHUser
	pwd := c.SSHPassword

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

	if (cmd == "ash") || cmd == ("bash") {
		err := attachShell(session)
		if err != nil {
			return err
		}
	} else {
		session.Stdout = os.Stdout
		session.Stderr = os.Stdout
		session.Stdin = os.Stdin

		if err := session.Run(cmd); err != nil {
			return err
		}
	}

	return nil

}

func attachShell(session *ssh.Session) error {
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal make raw: %s", err)
	}
	defer term.Restore(fd, state)

	width, height, err := term.GetSize(0)
	if err != nil {
		return err
	}

	terminal := os.Getenv("TERM")
	if terminal == "" {
		terminal = "xterm-256color"
	}

	if err := session.RequestPty(terminal, height, width, modes); err != nil {
		return err
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	if err := session.Shell(); err != nil {
		return err
	}

	if err := session.Wait(); err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			switch e.ExitStatus() {
			case 130:
				return nil
			}
		}
		return fmt.Errorf("ssh: %s", err)
	}

	return nil
}

// Status returns VM status
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

// Stop stops an Alpine VM
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

func getHostArchitecture() (string, error) {
	out, err := exec.Command("uname", "-m").Output()
	return strings.TrimSpace(string(out)), err
}

// IsNativeArch tests if VM architecture is the same as host
func (c *MachineConfig) IsNativeArch() bool {
	hostArch, err := getHostArchitecture()
	if err != nil {
		return false
	}

	return (hostArch == "arm64" && c.Arch == "aarch64") || (hostArch == "x86_64" && c.Arch == "x86_64")
}

// GetAccel Returns platform-appropriate accelerator
func (c *MachineConfig) GetAccel() string {
	if c.IsNativeArch() {
		switch runtime.GOOS {
		case "darwin":
			return "hvf"
		case "linux":
			return "kvm"
		case "windows":
			return "whpx" // untested
		}
	}
	return "tcg,tb-size=1024,thread=multi"
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

	qemuCmd := "qemu-system-" + c.Arch

	var qemuArgs []string

	cpuType := map[string]string{
		"aarch64": "cortex-a72",
		"x86_64":  "qemu64"}
	cpu := cpuType[c.Arch]

	aarch64Args := []string{
		"-M", "virt,highmem=off",
		"-bios", filepath.Join(c.Location, "qemu_efi.fd")}

	x86Args := []string{
		"-global", "PIIX4_PM.disable_s3=1",
		"-global", "ICH9-LPC.disable_s3=1",
	}

	mountArgs := []string{"-fsdev", "local,path=" + c.Mount + ",security_model=none,id=host0",
		"-device", "virtio-9p-pci,fsdev=host0,mount_tag=host0"}

	commonArgs := []string{
		"-m", c.Memory,
		"-cpu", cpu,
		"-accel", c.GetAccel(),
		"-smp", "cpus=" + c.CPU + ",sockets=1,cores=" + c.CPU + ",threads=1",
		"-drive", "if=virtio,file=" + filepath.Join(c.Location, c.Image),
		"-nographic",
		"-netdev", exposedPorts,
		"-device", "e1000,netdev=net0,mac=" + c.MACAddress,
		"-pidfile", filepath.Join(c.Location, "alpine.pid"),
		"-chardev", "socket,id=char-serial,path=" + filepath.Join(c.Location,
			"alpine.sock") + ",server=on,wait=off,logfile=" + filepath.Join(c.Location, "alpine.log"),
		"-serial", "chardev:char-serial",
		"-chardev", "socket,id=char-qmp,path=" + filepath.Join(c.Location, "alpine.qmp") + ",server=on,wait=off",
		"-qmp", "chardev:char-qmp",
		"-parallel", "none",
		"-device", "virtio-rng-pci",
		"-rtc", "base=utc,clock=host",
		"-name", c.Alias}

	if c.Arch == "aarch64" {
		qemuArgs = append(aarch64Args, commonArgs...)
	}
	if c.Arch == "x86_64" {
		qemuArgs = append(x86Args, commonArgs...)
	}

	if c.Mount != "" {
		qemuArgs = append(qemuArgs, mountArgs...)
	}

	cmd := exec.Command(qemuCmd, qemuArgs...)

	cmd.Stdout = os.Stdout

	// Uncomment to debug qemu messages
	cmd.Stderr = os.Stderr

	log.Printf("Booting...")
	err := cmd.Start()
	if err != nil {
		return err
	}

	err = utils.Retry(10, 2*time.Second, func() error {
		err := c.Exec("hwclock -s")
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.New("unable to sync clocks: " + err.Error())
	}

	if c.Mount != "" {
		err = utils.Retry(10, 2*time.Second, func() error {
			err := c.Exec("DIR=$(eval echo \"~$USER\");mkdir -p $DIR/mnt/; mount -t 9p -o trans=virtio host0 $DIR/mnt/ -oversion=9p2000.L,msize=104857600")
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return errors.New("unable to mount: " + err.Error())
		}

		log.Printf("Mounted " + c.Mount + " --> $HOME/mnt/")
	}

	status, _ := c.Status()

	if status == "Stopped" {
		return errors.New("unable to start vm")
	}

	return nil
}

// Launch macpine downloads a fresh image and creates a VM directory
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

	imageURL := utils.GetImageURL(c.Image)

	if _, err := os.Stat(filepath.Join(cacheDir, c.Image)); errors.Is(err, os.ErrNotExist) {
		err = utils.DownloadFile(filepath.Join(cacheDir, c.Image), imageURL)
		if err != nil {
			return errors.New("unable to download " + c.Image + " for " + c.Arch + ": " + err.Error())
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

	_, err = utils.CopyFile(filepath.Join(cacheDir, c.Image), filepath.Join(targetDir, c.Image))
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

	err = c.Start()
	if err != nil {
		return errors.New("unable launch a new machine. " + err.Error())
	}

	//Resize disk on an alpine guest
	if strings.Split(c.Image, "_")[0] == "alpine" {
		err = utils.Retry(10, 1*time.Second, func() error {
			err := c.Exec("apk add --no-cache e2fsprogs-extra sfdisk partx")
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return errors.New("unable install dependencies: " + err.Error())
		}

		err = utils.Retry(3, 1*time.Second, func() error {
			err = c.Exec(`echo ", +" | sfdisk --no-reread -N 3 /dev/vda; partx -u /dev/vda`)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return errors.New("error expanding disk: " + err.Error())
		}

		err = utils.Retry(3, 1*time.Second, func() error {
			err = c.Exec("resize2fs /dev/vda3")
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return errors.New("error expanding disk: " + err.Error())
		}

		err = c.Exec("df -h")
		if err != nil {
			return err
		}
	}

	return nil
}

// ResizeQemuDiskImage resizes a qcow2 disk image
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

// CreateQemuDiskImage creates a qcow2 disk image
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
