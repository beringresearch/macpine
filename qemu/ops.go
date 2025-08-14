package qemu

import (
	"bytes"
	"errors"
	"fmt"
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
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

type MachineConfig struct {
	Alias        string   `yaml:"alias"`
	Image        string   `yaml:"image"`
	Arch         string   `yaml:"arch"`
	CPU          string   `yaml:"cpu"`
	Memory       string   `yaml:"memory"`
	Disk         string   `yaml:"disk"`
	Mount        string   `yaml:"mount"`
	MachineIP    string   `yaml:"machineip"`
	Port         string   `yaml:"port"`
	VMNet        bool     `yaml:"vmnet"`
	SSHPort      string   `yaml:"sshport"`
	SSHUser      string   `yaml:"sshuser"`
	SSHPassword  string   `yaml:"sshpassword"`
	RootPassword *string  `yaml:"rootpassword,omitempty"`
	MACAddress   string   `yaml:"macaddress"`
	Location     string   `yaml:"location"`
	Tags         []string `yaml:"tags"`
}

// Exec starts an interactive shell terminal in VM
func (c *MachineConfig) Exec(cmd string, root bool) (string, error) {
	if cmd == "" {
		return "", nil
	}
	ip := c.MachineIP

	if c.VMNet {
		if ip == "localhost" || ip == "" {
			log.Println("getting instance IP address from DHCP leases")
			for {
				dhcpLeasesContent, err := os.ReadFile("/var/db/dhcpd_leases")
				if err != nil {
					return "", err
				}
				lip, _ := c.GetIPAddressByMac(dhcpLeasesContent)
				//ip = c.GetIPAddressFromMachine()
				if lip != "" {
					c.MachineIP = lip
					break
				}
				fmt.Print(".")
				time.Sleep(4 * time.Second)
			}

			config, err := yaml.Marshal(&c)

			if err != nil {
				c.Stop()
				c.CleanPIDFile()
				return "", err
			}

			err = os.WriteFile(filepath.Join(c.Location, "config.yaml"), config, 0644)
			if err != nil {
				c.Stop()
				c.CleanPIDFile()
				return "", err
			}
		}

	}

	host := c.MachineIP + ":" + c.SSHPort
	user := c.SSHUser
	pwd := c.SSHPassword

	if root && user != "root" {
		user = "root"
		if c.RootPassword == nil {
			pwd = "root"
		} else {
			pwd = *c.RootPassword
		}
	}
	cred, err := utils.GetCredential(pwd)
	if err != nil {
		return "", err
	}

	var conf *ssh.ClientConfig
	if cred.CRType == utils.PwdCred {
		conf = &ssh.ClientConfig{
			User:            user,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth: []ssh.AuthMethod{
				ssh.Password(cred.CR),
			},
		}
	} else { // utils.HostCred
		// Use SSH agent (https://pkg.go.dev/golang.org/x/crypto/ssh/agent#example-NewClient)
		socket := os.Getenv("SSH_AUTH_SOCK")
		conn, err := net.Dial("unix", socket)
		if err != nil {
			log.Fatalf("failed to open SSH_AUTH_SOCK: %v", err)
		}
		agentClient := agent.NewClient(conn)
		conf = &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeysCallback(agentClient.Signers),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}

	var conn *ssh.Client

	for i := 0; i < 11; i++ {
		conn, err = ssh.Dial("tcp", host, conf)
		if err == nil {
			break
		}
		if i == 10 {
			return "", err
		}

		fmt.Print(".")
		time.Sleep(4 * time.Second)
	}

	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var stdinBuf bytes.Buffer

	// XXX get shells from /etc/shells instead?
	if (cmd == "ash") || (cmd == "bash") {
		err := attachShell(session)
		if err != nil {
			return "", err
		}
	} else {

		session.Stdout = &stdoutBuf
		session.Stderr = &stderrBuf
		session.Stdin = &stdinBuf

		for i := 0; i < 5; i++ {
			err := session.Run(cmd)
			if err == nil {
				break
			}
			if i == 4 {
				return "", err
			}
		}
	}

	output := stdoutBuf.String()
	return output, nil
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
	status := "Stopped"
	var pid int

	pidFile := filepath.Join(c.Location, "alpine.pid")

	if _, err := os.Stat(pidFile); err == nil {
		status = "Running"
		pid, _ = c.GetInstancePID()

		// check if stopped and return "Paused"
		execArgs := []string{"-o", "stat=", "-p", strconv.Itoa(pid)}
		execCmd := "ps"

		cmd := exec.Command(execCmd, execArgs...)

		out, _ := cmd.Output()
		// if err != nil {
		// 	log.Fatalf("error checking status of qemu process: %v\n", err)
		// }
		if strings.TrimSpace(string(out)) == "T" {
			status = "Paused"
		}
	}

	return status, pid
}

// Stop stops an Alpine VM
func (c *MachineConfig) Stop() error {
	// qemu creates PID file with -pidfile flag, and deletes it on sigterm
	if status, pid := c.Status(); status != "Stopped" {
		if pid > 0 {
			p, procErr := os.FindProcess(pid)
			if procErr != nil {
				return procErr
			}

			if err := p.Signal(syscall.SIGKILL); err != nil {
				return err
			}

			pidFile := filepath.Join(c.Location, "alpine.pid")
			sockFile := filepath.Join(c.Location, "alpine.sock")
			qmpFile := filepath.Join(c.Location, "alpine.qmp")
			os.Remove(pidFile)
			os.Remove(sockFile)
			os.Remove(qmpFile)

			log.Println(c.Alias + " stopped")
			return nil
		} else {
			//pidFile := filepath.Join(c.Location, "alpine.pid")
			return errors.New("failed to stop instance `" + c.Alias + "` due to inadequate privileges")
		}
	}
	return nil
}

// Pauses an Alpine VM
func (c *MachineConfig) Pause() error {
	if status, pid := c.Status(); status == "Running" {
		if pid > 0 {
			p, procErr := os.FindProcess(pid)
			if procErr != nil {
				return procErr
			}
			if err := p.Signal(syscall.SIGSTOP); err != nil {
				return err
			}
			log.Println(c.Alias + " paused")
			return nil
		} else {
			pidFile := filepath.Join(c.Location, "alpine.pid")
			return errors.New("error pausing, incorrect PID in " + pidFile + "?")
		}
	}
	return nil
}

// Unpauses an Alpine VM
func (c *MachineConfig) Resume() error {
	if status, pid := c.Status(); status == "Paused" {
		if pid > 0 {
			p, procErr := os.FindProcess(pid)
			if procErr != nil {
				return procErr
			}
			if err := p.Signal(syscall.SIGCONT); err != nil {
				return err
			}
			_, err := c.Exec("hwclock -s", true)
			if err != nil {
				log.Println("failed to synchonrize clock, instance system clock may be skewed")
				return err
			}
			log.Println(c.Alias + " resumed")
			return nil
		} else {
			pidFile := filepath.Join(c.Location, "alpine.pid")
			return errors.New("error resuming, incorrect PID in " + pidFile + "?")
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

func (c *MachineConfig) HasHostCPU() bool {
	switch runtime.GOOS {
	case "darwin", "linux":
		return true
	case "netbsd", "windows":
		return false
	}
	// Not reached
	return false
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
	log.Println("Note: defaulting to QEMU tiny codegen. Emulation overhead may be significant.")
	return "tcg,tb-size=1024,thread=multi"
}

// Start starts up an Alpine VM
func (c *MachineConfig) Start() error {

	networkDevice := "user,id=net0,hostfwd=tcp::" + c.SSHPort + "-:22"

	if c.MACAddress == "" {
		macAddress, err := utils.GenerateMACAddress()
		if err != nil {
			log.Fatal(err)
		}

		c.MACAddress = macAddress
	}

	if c.VMNet {
		networkDevice = "vmnet-shared,id=net0"
	}

	// Only parse ports of using qemu's default slirp network
	if !c.VMNet {
		ports, err := utils.ParsePort(c.Port)
		if err != nil {
			log.Fatalf("Error configuring ports: %v\n", err)
		}
		for _, p := range ports {
			hostp := strconv.Itoa(p.Host)
			guestp := strconv.Itoa(p.Guest)
			if p.Proto == utils.Tcp {
				networkDevice += ",hostfwd=tcp::" + hostp + "-:" + guestp
			} else { // Udp
				networkDevice += ",hostfwd=udp::" + hostp + "-:" + guestp
			}
		}
	}

	qemuCmd := "qemu-system-" + c.Arch

	var qemuArgs []string

	hostCPU := ""

	if c.IsNativeArch() {
		if c.HasHostCPU() {
			hostCPU = "_host"
		}
	}

	hostCPUType := c.Arch + hostCPU
	hugePages := ""

	if hostCPUType == "x86_64_host" {
		supports, err := utils.SupportsHugePages()
		if err != nil {
			return err
		}

		if !supports {
			hugePages = ",pdpe1gb=off"
		}
	}

	cpuType := map[string]string{
		"aarch64":      "cortex-a72",
		"x86_64":       "max,+avx,+avx2",
		"x86_64_host":  "host" + hugePages,
		"aarch64_host": "host"}

	cpu := cpuType[c.Arch+hostCPU]

	highmem := "off"
	intMem, err := strconv.Atoi(c.Memory)
	if err != nil {
		return err
	}
	if intMem > 2000 {
		highmem = "on"
	}

	aarch64Args := []string{
		"-M", "virt,highmem=" + highmem,
		"-bios", filepath.Join(c.Location, "qemu_efi.fd"),
	}

	x86Args := []string{
		"-global", "PIIX4_PM.disable_s3=1",
		"-global", "ICH9-LPC.disable_s3=1",
	}

	mountArgs := []string{"-fsdev", "local,path=" + c.Mount + ",security_model=mapped-xattr,id=host0",
		"-device", "virtio-9p-pci,fsdev=host0,mount_tag=host0"}

	commonArgs := []string{
		"-m", c.Memory,
		"-cpu", cpu,
		"-accel", c.GetAccel(),
		"-smp", "cpus=" + c.CPU + ",sockets=1,cores=" + c.CPU + ",threads=1",
		"-drive", "if=virtio,file=" + filepath.Join(c.Location, c.Image),
		"-nographic",
		"-device", "virtio-net-pci,netdev=net0,mac=" + c.MACAddress,
		"-netdev", networkDevice,
		"-pidfile", filepath.Join(c.Location, "alpine.pid"),
		"-chardev", "socket,id=char-serial,path=" + filepath.Join(c.Location,
			"alpine.sock") + ",server=on,wait=off,logfile=" + filepath.Join(c.Location, "alpine.log"),
		"-serial", "chardev:char-serial",
		"-chardev", "socket,id=char-qmp,path=" + filepath.Join(c.Location, "alpine.qmp") + ",server=on,wait=off",
		"-qmp", "chardev:char-qmp",
		"-parallel", "none",
		"-device", "virtio-rng-pci",
		"-rtc", "base=utc,clock=host",
		"-daemonize",
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

	log.Println("booting " + c.Alias)
	err = cmd.Run()
	if err != nil {
		c.Stop()
		c.CleanPIDFile()
		return err
	}

	if c.Mount != "" {
		basename := filepath.Base(c.Mount)
		mntcmd := make([]string, 3)
		mntcmd[0] = "mkdir -p /mnt/" + basename
		mntcmd[1] = "chmod 777 /mnt/" + basename
		mntcmd[2] = "mount -t 9p -o trans=virtio,version=9p2000.L,msize=104857600 host0 /mnt/" + basename
		if _, err := c.Exec(strings.Join(mntcmd, " && "), true); err != nil {
			log.Println("error mounting directory: " + err.Error())
		} else {
			log.Println("mounted " + c.Mount + " on /mnt/" + basename)
		}
	}

	status, pid := c.Status()
	if status != "Running" {
		return errors.New("unable to start instance")
	}

	// err = c.Exec("hwclock -s", true)
	// if err != nil {
	// 	c.Stop()
	// 	c.CleanPIDFile()
	// 	return err
	// }

	config, err := yaml.Marshal(&c)
	if err != nil {
		c.Stop()
		c.CleanPIDFile()
		return err
	}

	err = os.WriteFile(filepath.Join(c.Location, "config.yaml"), config, 0644)
	if err != nil {
		c.Stop()
		c.CleanPIDFile()
		return err
	}

	log.Println(c.Alias + " started (" + strconv.Itoa(pid) + ")")

	return nil
}

// func (c *MachineConfig) GetIPAddressFromMachine() string {
// 	ip := ""

// 	cmd := "ip route get 1.2.3.4 | awk '{print $7}'"
// 	ip, err := c.Exec(cmd, false)
// 	if err != nil {
// 		return ""
// 	}

// 	fmt.Println(ip)

// 	return ip
// }

// AssignIP obtains machine IP address from bootpd.plist. Only applicale to machines created on
// VMNet
func (c *MachineConfig) GetIPAddressByMac(dhcpLeasesContent []byte) (string, error) {
	result := utils.ParseDhcpLeasesFile(string(dhcpLeasesContent))
	dhcpData := utils.ConvertStringArrayToDhcpDataArray(result)
	dhcpConfig := utils.MatchHwAddress(dhcpData, c.MACAddress)

	if dhcpConfig != nil {
		return dhcpConfig.IpAddress, nil
	} else {
		return "", errors.New("no machine dhcp configuration found")
	}
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

	err = os.WriteFile(filepath.Join(c.Location, "config.yaml"), config, 0644)
	if err != nil {
		os.RemoveAll(targetDir)
		return err
	}

	err = c.Start()
	if err != nil {
		return errors.New("unable to launch a new machine. " + err.Error())
	}

	log.Println("waiting for machine...")
	time.Sleep(10 * time.Second)

	// Make sure DNS is set up correctly
	_, err = c.Exec("echo 'nameserver 8.8.8.8' > /etc/resolv.conf", true)
	if err != nil {
		return errors.New("unable to set up DNS: " + err.Error())
	}

	_, err = c.Exec("apk update && apk add --no-cache dhclient", true)
	if err != nil {
		return errors.New("unable to install dhclient: " + err.Error())
	}

	_, err = c.Exec(`cat >/etc/dhcp/dhclient.conf <<EOL
	option rfc3442-classless-static-routes code 121 = array of unsigned integer 8;

	send host-name = gethostname();
	request subnet-mask, broadcast-address, time-offset, routers,
	        domain-name, domain-name-servers, domain-search, host-name,
	        dhcp6.name-servers, dhcp6.domain-search, dhcp6.fqdn, dhcp6.sntp-servers,
	        netbios-name-servers, netbios-scope, interface-mtu,
	        rfc3442-classless-static-routes, ntp-servers;

	prepend domain-name-servers 8.8.8.8, 8.8.4.4;
	EOL`, true)
	if err != nil {
		return errors.New("unable to configure dhclient: " + err.Error())
	}

	_, err = c.Exec("rc-service networking restart", true)
	if err != nil {
		return errors.New("unable to restart networking services: " + err.Error())
	}

	// Resize disk on an alpine guest
	if strings.Split(c.Image, "_")[0] == "alpine" {
		//TODO add these dependencies into pre-baked macpine image
		_, err := c.Exec("apk add --no-cache e2fsprogs-extra sfdisk partx", true) // root=true i.e. run as root
		if err != nil {
			return errors.New("unable to install dependencies: " + err.Error())
		}

		//send sfdisk command ,+ (<start>,<size>,<type>,<bootable>)
		//default start (0), size + (all available), default type (linux data), default bootable (false)
		_, err = c.Exec(`echo ",+" | sfdisk --no-reread --partno 3 /dev/vda && partx -u /dev/vda`, true)
		if err != nil {
			return errors.New("error updating partition table: " + err.Error())
		}

		_, err = c.Exec("resize2fs /dev/vda3", true)
		if err != nil {
			return errors.New("error expanding filesystem: " + err.Error())
		}

		_, err = c.Exec("df -h", true)
		if err != nil {

			return err
		}
	}

	return nil
}

// CompressQemuDiskImage compresses the QEMU Disk image and overwrites it
func (c *MachineConfig) CompressQemuDiskImage() error {
	if !utils.CommandExists("qemu-img") {
		return errors.New("qemu-img is not available on $PATH. ensure qemu is installed")
	}

	cmd := exec.Command("qemu-img", "convert", "-c", "-O", "qcow2", filepath.Join(c.Location, c.Image),
		filepath.Join(c.Location, c.Image+"_compressed.qcow2"))

	err := cmd.Run()
	if err != nil {
		return err
	}

	err = os.Remove(filepath.Join(c.Location, c.Image))
	if err != nil {
		return err
	}

	err = os.Rename(filepath.Join(c.Location, c.Image+"_compressed.qcow2"),
		filepath.Join(c.Location, c.Image))
	if err != nil {
		return err
	}

	return nil
}

// DecompressQemuDiskImage decompresses the QEMU Disk image and overwrites it
func (c *MachineConfig) DecompressQemuDiskImage() error {
	if !utils.CommandExists("qemu-img") {
		return errors.New("qemu-img is not available on $PATH. ensure qemu is installed")
	}

	cmd := exec.Command("qemu-img", "convert", "-O", "qcow2", "-p", filepath.Join(c.Location, c.Image),
		filepath.Join(c.Location, c.Image+"_decompressed.qcow2"))

	err := cmd.Run()
	if err != nil {
		return err
	}

	err = os.Remove(filepath.Join(c.Location, c.Image))
	if err != nil {
		return err
	}

	err = os.Rename(filepath.Join(c.Location, c.Image+"_decompressed.qcow2"),
		filepath.Join(c.Location, c.Image))
	if err != nil {
		return err
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

func (c *MachineConfig) CleanPIDFile() {
	pidFile := filepath.Join(c.Location, "alpine.pid")
	if err := os.Remove(pidFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("error deleting pidfile at %s. Manually delete it before proceeding.", pidFile)
	}
}

func (c *MachineConfig) GetInstancePID() (int, error) {
	pidFile := filepath.Join(c.Location, "alpine.pid")

	_, err := os.Open(pidFile)
	if err != nil {
		return 0, errors.New("you don't have enough priveledges to view " + c.Alias + ". run sudo alpine list")
	}

	vmPID, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	process := string(vmPID)
	process = strings.TrimSuffix(process, "\n")
	pid, _ := strconv.Atoi(process)

	return pid, nil
}

func GetMachineConfig(vmName string) (MachineConfig, error) {
	machineConfig := MachineConfig{}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return machineConfig, err
	}

	config, err := os.ReadFile(filepath.Join(userHomeDir, ".macpine", vmName, "config.yaml"))
	if err != nil {
		return machineConfig, err
	}

	err = yaml.Unmarshal(config, &machineConfig)
	if err != nil {
		return machineConfig, err
	}
	return machineConfig, nil
}

func SaveMachineConfig(machineConfig MachineConfig) error {
	updatedConfig, err := yaml.Marshal(&machineConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(machineConfig.Location, "config.yaml"), updatedConfig, 0644)
	if err != nil {
		return err
	}
	return nil
}
