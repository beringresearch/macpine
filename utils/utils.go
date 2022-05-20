package utils

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"
)

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
