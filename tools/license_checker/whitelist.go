package main

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
)

type whitelist []whitelistEntry

type whitelistEntry struct {
	Name     string
	Version  *semver.Version
	Operator string
	License  string
}

var re = regexp.MustCompile(`^([<=>]+\s*)?([-\.\/\w]+)(@v[\d\.]+)?\s+([-\.\w]+)$`)

func (w *whitelist) Parse(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read file line-by-line and split by semicolon
	lineno := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineno++
		if strings.HasPrefix(line, "#") {
			// Comment
			continue
		}

		groups := re.FindAllStringSubmatch(line, -1)
		if len(groups) != 1 {
			log.Printf("WARN: Ignoring not matching entry in line %d", lineno)
			continue
		}
		group := groups[0]
		if len(group) != 5 {
			// Malformed
			log.Printf("WARN: Ignoring malformed entry in line %d", lineno)
			continue
		}

		// An entry has the form:
		// [operator]<package name>[@version] [license SPDX]
		var operator, version string
		if group[1] != "" {
			operator = strings.TrimSpace(group[1])
		}
		name := group[2]
		if group[3] != "" {
			version = strings.TrimSpace(group[3])
			version = strings.TrimLeft(version, "@v")
		}
		license := strings.TrimSpace(group[4])

		entry := whitelistEntry{Name: name, License: license, Operator: operator}
		if version != "" {
			entry.Version, err = semver.NewVersion(version)
			if err != nil {
				// Malformed
				log.Printf("Ignoring malformed version in line %d: %v", lineno, err)
				continue
			}
			if entry.Operator == "" {
				entry.Operator = "="
			}
		}
		*w = append(*w, entry)
	}

	return scanner.Err()
}

func (w *whitelist) Check(pkg, version, spdx string) (ok, found bool) {
	var pkgver semver.Version
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "v")
	if v != "" {
		pkgver = *semver.New(v)
	}

	for _, entry := range *w {
		if entry.Name != pkg {
			continue
		}

		var match bool
		switch entry.Operator {
		case "":
			match = true
		case "=":
			match = pkgver.Equal(*entry.Version)
		case "<":
			match = pkgver.LessThan(*entry.Version)
		case "<=":
			match = pkgver.LessThan(*entry.Version) || pkgver.Equal(*entry.Version)
		case ">":
			match = !(pkgver.LessThan(*entry.Version) || pkgver.Equal(*entry.Version))
		case ">=":
			match = !pkgver.LessThan(*entry.Version)
		}
		if match {
			return entry.License == spdx, true
		}
	}

	return false, false
}
