package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

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

	adjectives, err := readLines("utils/adjectives.txt")
	if err != nil {
		log.Fatal(err)
	}

	nouns, err := readLines("utils/nouns.txt")
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(adjectives)
	a := adjectives[n]

	rand.Seed(time.Now().Unix())
	n = rand.Int() % len(nouns)
	o := nouns[n]

	alias = a + "-" + o
	return alias

}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
