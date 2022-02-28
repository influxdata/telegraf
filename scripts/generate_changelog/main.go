package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log" //nolint:revive
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Run git-chglog
	ver, err := os.ReadFile("build_version.txt")
	if err != nil {
		log.Fatal(err)
	}

	version := fmt.Sprintf("v%s", strings.TrimSuffix(string(ver), "\n"))

	out, err := exec.Command("git-chglog", "--next-tag", version, fmt.Sprintf("%s..", version)).Output()
	if err != nil {
		log.Fatal(err)
	}

	// Appned to exisiting CHANGELOG.md
	changelogFile, err := os.Open("CHANGELOG.md")
	if err != nil {
		log.Fatal(err)
	}
	defer changelogFile.Close()

	var c []byte
	buf := bytes.NewBuffer(c)

	scanner := bufio.NewScanner(changelogFile)
	var read bool
	for scanner.Scan() {
		if !read && scanner.Text() == "# Changelog" {
			read = true
			continue
		}

		if read {
			_, err := buf.Write(scanner.Bytes())
			if err != nil {
				log.Fatal(err)
			}
			_, err = buf.WriteString("\n")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	header := `<!-- markdownlint-disable MD024 -->

# Changelog
`
	out = append([]byte(header)[:], out[:]...)
	out = append(out[:], buf.Bytes()[:]...)

	err = os.WriteFile("CHANGELOG_NEW.md", out, 0664)
	if err != nil {
		log.Fatal(err)
	}
}
