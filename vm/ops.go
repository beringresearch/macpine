package vm

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/beringresearch/macpine/utils"
)

type MachineConfig struct {
	Alias    string
	Arch     string
	Version  string
	CPU      string
	Memory   string
	Disk     string
	Port     string
	Location string
}

func (c *MachineConfig) Install() error {

	imageName, _ := utils.GetAlpineURL(c.Version, c.Arch)

	cmd := exec.Command("qemu-system-aarch64",
		"-M", "virt,highmem=off",
		"-m", "2048",
		"-rtc", "base=utc,clock=host,driftfix=slew",
		"-bios", filepath.Join(c.Location, "QEMU_EFI.fd"),
		"-device", "virtio-rng-pci",
		"-device", "virtio-balloon",
		"-nographic", "-no-reboot",
		"-serial", "mon:stdio",
		"-drive", "if=virtio,file="+filepath.Join(c.Location, "user-data.qcow2"),
		"-monitor", "unix:monitor.sock,server,nowait",
		"-netdev", "user,id=net1,hostfwd=tcp:127.0.0.1:"+c.Port+"-:22",
		"-device", "virtio-net-pci,netdev=net1",
		"-smp", "4",
		"-cdrom", filepath.Join(c.Location, imageName),
		"-drive", "if=virtio,file="+filepath.Join(c.Location, "tmp.qcow2"),
		"-cpu", "host",
		"-accel", "hvf")

	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		return errors.New("cmd.Start() failed with " + err.Error())
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		_, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		return errors.New("cmd.Run() failed with " + err.Error())
	}
	if errStdout != nil || errStderr != nil {
		return errors.New("failed to capture stdout or stderr")
	}

	return nil
}

//Initialise macpine if this is a fresh install
func (c *MachineConfig) Init() error {
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
		log.Println("Downloading " + imageName)
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

	err = utils.DownloadFile(filepath.Join(targetDir, "QEMU_EFI.fd"),
		"http://releases.linaro.org/components/kernel/uefi-linaro/16.02/release/qemu64/QEMU_EFI.fd")
	if err != nil {
		return errors.New("unable to download Alpine " + c.Version + " for " + c.Arch + ": " + err.Error())
	}

	_, err = utils.CopyFile(filepath.Join(cacheDir, imageName), filepath.Join(targetDir, imageName))
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

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)

			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
