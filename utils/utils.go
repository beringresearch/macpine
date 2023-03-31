package utils

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//go:embed *.txt
var f embed.FS

// GenerateMACAddress
func GenerateMACAddress() (string, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	// Set the local bit
	buf[0] |= 2
	mac := fmt.Sprintf("56:%02x:%02x:%02x:%02x:%02x", buf[1], buf[2], buf[3], buf[4], buf[5])

	return mac, nil
}

// Retry retries a function
func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			fmt.Printf("\r%s", strings.Repeat(".", i))
			time.Sleep(sleep)
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

type Protocol int

const (
	Tcp Protocol = iota
	Udp
)

type PortMap struct {
	Host  int
	Guest int
	Proto Protocol
}

// Parses port mapping configurations
func ParsePort(ports string) ([]PortMap, error) {
	var maps []PortMap = nil
	if ports == "" {
		return maps, nil
	}
	mapcount := strings.Count(ports, ",") + 1
	if mapcount > 65535 {
		return nil, errors.New("Too many port mappings specified, likely an error. Check config.yaml")
	}
	maps = make([]PortMap, mapcount)
	for i, p := range strings.Split(ports, ",") {
		newmap := PortMap{0, 0, Tcp}
		var herr, gerr error = nil, nil
		if strings.HasSuffix(p, "u") {
			newmap.Proto = Udp
			p = strings.TrimSuffix(p, "u")
		}
		if strings.Contains(p, ":") {
			pair := strings.Split(p, ":")
			if len(pair) != 2 {
				return nil, errors.New("Incorrect port mapping pair specified. Check config.yaml")
			}
			newmap.Host, herr = strconv.Atoi(pair[0])
			newmap.Guest, gerr = strconv.Atoi(pair[1])
		} else {
			newmap.Host, herr = strconv.Atoi(p)
			newmap.Guest = newmap.Host
		}
		if herr != nil || gerr != nil {
			return nil, errors.New("Error parsing specified ports. Check config.yaml")
		}
		if newmap.Host < 0 || newmap.Host > 65535 || newmap.Guest < 0 || newmap.Guest > 65535 {
			return nil, errors.New("Invalid specified ports (must be 0-65535). Check config.yaml")
		}
		maps[i] = newmap
	}
	return maps, nil
}

// Ping checks if connection is reachable
func Ping(ip string, port string) error {
	address, err := net.ResolveTCPAddr("tcp", ip+":"+port)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		return nil
	}

	if conn != nil {
		defer conn.Close()
		return errors.New("port " + port + " already assigned on host")
	}

	return err
}

// StringSliceContains check if string value is in []string
func StringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Uncompress uncompresses gzip
func Uncompress(source string, destination string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	gzRead, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzRead.Close()

	tarRead := tar.NewReader(gzRead)
	for {
		cur, err := tarRead.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		os.MkdirAll(destination, 0777)

		switch cur.Typeflag {

		case tar.TypeReg:
			create, err := os.Create(filepath.Join(destination, cur.Name))
			if err != nil {
				return err
			}
			defer create.Close()
			create.ReadFrom(tarRead)
		case tar.TypeLink:
			os.Link(cur.Linkname, cur.Name)
		}
	}
	return nil
}

// Compress creates a tar.gz of a Directory
func Compress(files []string, buf io.Writer) error {
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

// CopyFile copies file from src to dst
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rRetrieving image... %3dMB complete", wc.Total/1000000)
}

func DownloadFile(filepath string, url string) error {
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		out.Close()
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("requested image download is not supported: StatusCode " + strconv.Itoa(resp.StatusCode))
	}

	defer resp.Body.Close()

	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	fmt.Print("\n")
	out.Close()

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}
	return nil
}

func GetImageURL(version string) string {
	url := "https://github.com/beringresearch/macpine/releases/download/v.01/" + version
	return url
}

func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func GenerateRandomAlias() string {
	var alias string
	var adjectivesString []string
	var nounsString []string

	adjectives, _ := f.ReadFile("adjectives.txt")
	adjectivesString = strings.Split(string(adjectives), "\n")

	nouns, _ := f.ReadFile("nouns.txt")
	nounsString = strings.Split(string(nouns), "\n")

	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(adjectivesString)
	a := adjectivesString[n]

	rand.Seed(time.Now().Unix())
	n = rand.Int() % len(nounsString)
	o := nounsString[n]

	alias = a + "-" + o
	return alias

}

type CmdResult struct {
   Name string;
   Err error;
}
