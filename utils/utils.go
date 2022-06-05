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
	"strings"
	"time"
)

//go:embed *.txt
var f embed.FS

//Ping checks if connection is reachable
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

//StringSliceContains check if string value is in []string
func StringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//Uncompress uncompresses gzip
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

//CopyFile copies file from src to dst
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

func GetAlpineURL(version string, arch string) (string, string) {
	imageFile := "alpine_" + version + "-" + arch + ".qcow2"
	url := "https://github.com/beringresearch/macpine/releases/download/v.01/" + imageFile
	return imageFile, url
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
