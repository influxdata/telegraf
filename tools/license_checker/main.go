package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log" //nolint:revive // We cannot use the Telegraf's logging here
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/mod/modfile"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

//go:embed data/spdx_mapping.json
var spdxMappingFile []byte

var debug bool
var spdxCache *cache
var nameToSPDX map[string]string

func debugf(format string, v ...any) {
	if !debug {
		return
	}
	log.Printf("DEBUG: "+format, v...)
}

func main() {
	var help, verbose bool
	var threshold float64
	var cacheFn, whitelistFn, userpkg string
	var expiry time.Duration

	flag.BoolVar(&debug, "debug", false, "output debugging information")
	flag.BoolVar(&help, "help", false, "output this help text")
	flag.BoolVar(&verbose, "verbose", false, "output verbose information instead of just errors")
	flag.Float64Var(&threshold, "threshold", 0.8, "threshold for license classification")
	flag.StringVar(&cacheFn, "cache", "", "use the given cache file")
	flag.StringVar(&whitelistFn, "whitelist", "", "use the given white-list file for comparison")
	flag.StringVar(&userpkg, "package", "", "only test the given package (all by default)")
	flag.DurationVar(&expiry, "expiry", 0, "time until a cache entry expires (never by default)")
	flag.Parse()

	if help || flag.NArg() > 1 {
		//nolint:revive // We cannot do anything about possible failures here
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [options] [telegraf root dir]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Arguments:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  telegraf root dir (optional)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "		path to the root directory of telegraf (default: .)\n")
		os.Exit(1)
	}

	// Setup full-name to license SPDX identifier mapping
	if err := json.Unmarshal(spdxMappingFile, &nameToSPDX); err != nil {
		log.Fatalf("Unmarshalling license name to SPDX mapping failed: %v", err)
	}

	// Get required files
	path := "."
	if flag.NArg() == 1 {
		path = flag.Arg(0)
	}

	moduleFilename := filepath.Join(path, "go.mod")
	licenseFilename := filepath.Join(path, "docs", "LICENSE_OF_DEPENDENCIES.md")

	var override whitelist
	if whitelistFn != "" {
		log.Printf("Reading whitelist file %q...", whitelistFn)
		if err := override.Parse(whitelistFn); err != nil {
			log.Fatalf("Reading whitelist failed: %v", err)
		}
	}

	if cacheFn != "" {
		var err error

		log.Printf("Reading cache file %q...", cacheFn)
		spdxCache, err = LoadCache(cacheFn)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Fatalf("Reading cache file failed: %v", err)
			}
			spdxCache = NewCache(expiry)
		}
		spdxCache.Expiry = expiry
	}

	log.Printf("Reading module file %q...", moduleFilename)
	modbuf, err := os.ReadFile(moduleFilename)
	if err != nil {
		log.Fatal(err)
	}
	depModules, err := modfile.Parse(moduleFilename, modbuf, nil)
	if err != nil {
		log.Fatalf("Parsing modules failed: %f", err)
	}
	debugf("found %d required packages", len(depModules.Require))

	dependencies := make(map[string]string)
	for _, d := range depModules.Require {
		dependencies[d.Mod.Path] = d.Mod.Version
	}

	log.Printf("Reading license file %q...", licenseFilename)
	licensesMarkdown, err := os.ReadFile(licenseFilename)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the markdown document
	parser := goldmark.DefaultParser()
	root := parser.Parse(text.NewReader(licensesMarkdown))

	// Prepare a line parser
	lineParser := goldmark.DefaultParser()

	// Collect the licenses
	// For each list we search for the items and parse them.
	// Expect a pattern of <package name> <link>.
	ignored := 0
	var packageInfos []packageInfo
	for node := root.FirstChild(); node != nil; node = node.NextSibling() {
		listNode, ok := node.(*ast.List)
		if !ok {
			continue
		}

		for inode := listNode.FirstChild(); inode != nil; inode = inode.NextSibling() {
			itemNode, ok := inode.(*ast.ListItem)
			if !ok || itemNode.ChildCount() != 1 {
				continue
			}
			textNode, ok := itemNode.FirstChild().(*ast.TextBlock)
			if !ok || textNode.Lines().Len() != 1 {
				continue
			}

			lineSegment := textNode.Lines().At(0)
			line := lineSegment.Value(licensesMarkdown)
			lineRoot := lineParser.Parse(text.NewReader(line))
			if lineRoot.ChildCount() != 1 || lineRoot.FirstChild().ChildCount() < 2 {
				log.Printf("WARN: Ignoring item %q due to wrong count (%d/%d)", string(line), lineRoot.ChildCount(), lineRoot.FirstChild().ChildCount())
				ignored++
				continue
			}

			var name, license, link string
			for lineElementNode := lineRoot.FirstChild().FirstChild(); lineElementNode != nil; lineElementNode = lineElementNode.NextSibling() {
				switch v := lineElementNode.(type) {
				case *ast.Text:
					name += string(v.Text(line))
				case *ast.Link:
					license = string(v.Text(line))
					link = string(v.Destination)
				default:
					debugf("ignoring unknown element %T (%v)", v, v)
				}
			}

			info := packageInfo{
				name:    strings.TrimSpace(name),
				version: dependencies[name],
				url:     strings.TrimSpace(link),
				license: strings.TrimSpace(license),
			}
			info.ToSPDX()
			if info.name == "" {
				log.Printf("WARN: Ignoring item %q due to empty package name", string(line))
				ignored++
				continue
			}
			if info.url == "" {
				log.Printf("WARN: Ignoring item %q due to empty url name", string(line))
				ignored++
				continue
			}
			if info.license == "" {
				log.Printf("WARN: Ignoring item %q due to empty license name", string(line))
				ignored++
				continue
			}
			debugf("adding %q with license %q (%s) and version %q at %q...", info.name, info.license, info.spdx, info.version, info.url)
			packageInfos = append(packageInfos, info)
		}
	}

	// Get the superset of licenses
	if debug {
		licenseSet := map[string]bool{}
		licenseNames := []string{}
		for _, info := range packageInfos {
			if found := licenseSet[info.license]; !found {
				licenseNames = append(licenseNames, info.license)
			}
			licenseSet[info.license] = true
		}
		sort.Strings(licenseNames)
		log.Println("Using licenses:")
		for _, license := range licenseNames {
			log.Println("  " + license)
		}
	}

	// Check the licenses by matching their text and compare the classification result
	// with the information provided by the user
	var succeeded, warn, failed int
	for _, info := range packageInfos {
		// Ignore all packages except the ones given by the user (if any)
		if userpkg != "" && userpkg != info.name {
			continue
		}

		// Check if we got a whitelist entry for the package
		if ok, found := override.Check(info.name, info.version, info.spdx); found {
			if ok {
				log.Printf("OK: %q (%s) (whitelist)", info.name, info.license)
				succeeded++
			} else {
				log.Printf("ERR: %q (%s) %s does not match whitelist", info.name, info.license, info.spdx)
				failed++
			}
			continue
		}

		// Perform a text classification
		confidence, err := info.Classify()
		if err != nil {
			log.Printf("ERR: %q (%s) %v", info.name, info.license, err)
			failed++
			continue
		}
		if confidence < threshold {
			log.Printf("WARN: %q (%s) has low matching confidence (%.2f%%)", info.name, info.license, confidence)
			warn++
			continue
		}
		if verbose {
			log.Printf("OK: %q (%s) (%.2f%%)", info.name, info.license, confidence)
		}
		succeeded++
	}
	if verbose {
		log.Printf("Checked %d licenses (%d ignored lines):", len(packageInfos), ignored)
		log.Printf("    %d successful", succeeded)
		log.Printf("    %d low confidence", warn)
		log.Printf("    %d errors", failed)
	}

	if cacheFn != "" {
		log.Printf("Writing cache file %q...", cacheFn)
		if err := spdxCache.Save(cacheFn); err != nil {
			log.Fatalf("Writing cache file failed: %v", err)
		}
	}

	if failed > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
