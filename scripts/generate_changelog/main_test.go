package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseCommits(t *testing.T) {
	ver, err := os.ReadFile("../../build_version.txt")
	if err != nil {
		log.Fatal(err)
	}

	version := fmt.Sprintf("v%s", strings.TrimSuffix(string(ver), "\n"))

	commits, err := ParseCommits()
	if err != nil {
		log.Fatal(err)
	}

	commitGroups := CreateCommitGroups(commits)

	newChanges := NewChanges{
		Version:      version,
		Date:         time.Now().Format("2006-01-02"),
		CommitGroups: commitGroups,
	}

	temp := template.Must(template.ParseFiles("scripts/generate_changelog/CHANGELOG.go.tmpl"))
	var out bytes.Buffer
	err = temp.Execute(&out, newChanges)
	if err != nil {
		log.Fatal(err)
	}

	err = AppendToChangelog(out.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}
