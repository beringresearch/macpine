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

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func GetAlpineURL(version string, arch string) (string, string) {
	imageFile := "alpine-standard-" + version + "-" + arch + ".iso"

	shortVersion := strings.Split(version, ".")

	url := "https://dl-cdn.alpinelinux.org/alpine/v" + strings.Join(shortVersion[0:2], ".") + "/releases/" + arch + "/" + imageFile
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
