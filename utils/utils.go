package utils

import (
	"embed"
	_ "embed"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:embed *.txt
var f embed.FS

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
	fmt.Printf("\rDownloading... %3dMB complete", wc.Total/1000000)
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
