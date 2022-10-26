// This is a tool to embed configuration files into the README.md of all plugins
// It searches for TOML sections in the plugins' README.md and detects includes specified in the form
//
//	```toml [@includeA.conf[ @includeB[ @...]]
//	    Whatever is in here gets replaced.
//	```
//
// Then it will replace everything in this section by the concatenation of the file `includeA.conf`, `includeB` etc.
// content. The tool is not stateful, so it can be run multiple time with a stable result as long
// as the included files do not change.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var (
	// Finds all comment section parts `<-- @includefile -->`
	commentIncludesEx = regexp.MustCompile(`<!--\s+(@.+)+\s+-->`)
	// Finds all TOML sections of the form `toml @includefile`
	tomlIncludesEx = regexp.MustCompile(`[\s"]+(@.+)+"?`)
	// Extracts the `includefile` part
	includeMatch = regexp.MustCompile(`(?:@([^\s"]+))+`)
)

type includeBlock struct {
	Includes []string
	Start    int
	Stop     int
	Newlines bool
}

func extractIncludeBlock(txt []byte, includesEx *regexp.Regexp, root string) *includeBlock {
	includes := includesEx.FindSubmatch(txt)
	if len(includes) != 2 {
		return nil
	}
	block := includeBlock{}
	for _, inc := range includeMatch.FindAllSubmatch(includes[1], -1) {
		if len(inc) != 2 {
			continue
		}
		include := filepath.FromSlash(string(inc[1]))
		// Make absolute paths relative to the include-root if any
		if filepath.IsAbs(include) {
			if root == "" {
				log.Printf("Ignoring absolute include %q without include root...", include)
				continue
			}
			include = filepath.Join(root, include)
		}
		include, err := filepath.Abs(include)
		if err != nil {
			log.Printf("Cannot resolve include %q...", include)
			continue
		}
		if fi, err := os.Stat(include); err != nil || !fi.Mode().IsRegular() {
			log.Printf("Ignoring include %q as it cannot be found or is not a regular file...", include)
			continue
		}
		block.Includes = append(block.Includes, include)
	}
	return &block
}

func insertInclude(buf *bytes.Buffer, include string) error {
	file, err := os.Open(include)
	if err != nil {
		return fmt.Errorf("opening include %q failed: %v", include, err)
	}
	defer file.Close()

	// Write the include and make sure we get a newline
	if _, err := io.Copy(buf, file); err != nil {
		return fmt.Errorf("inserting include %q failed: %v", include, err)
	}
	return nil
}

func insertIncludes(buf *bytes.Buffer, b *includeBlock) error {
	// Insert newlines before and after
	if b.Newlines {
		if _, err := buf.Write([]byte("\n")); err != nil {
			return errors.New("adding newline failed")
		}
	}

	// Insert all includes in the order they occurred
	for _, include := range b.Includes {
		if err := insertInclude(buf, include); err != nil {
			return err
		}
	}
	// Make sure we add a trailing newline
	if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		if _, err := buf.Write([]byte("\n")); err != nil {
			return errors.New("adding newline failed")
		}
	}

	// Insert newlines before and after
	if b.Newlines {
		if _, err := buf.Write([]byte("\n")); err != nil {
			return errors.New("adding newline failed")
		}
	}

	return nil
}

func main() {
	// Estimate Telegraf root to be able to handle absolute paths
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot get working directory: %v", err)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		log.Fatalf("Cannot resolve working directory: %v", err)
	}

	var includeRoot string
	if idx := strings.LastIndex(cwd, filepath.FromSlash("/plugins/")); idx > 0 {
		includeRoot = cwd[:idx]
	}

	// Get the file permission of the README for later use
	inputFilename := "README.md"
	inputFileInfo, err := os.Lstat(inputFilename)
	if err != nil {
		log.Fatalf("Cannot get file permissions: %v", err)
	}
	perm := inputFileInfo.Mode().Perm()

	// Read and parse the README markdown file
	readme, err := os.ReadFile(inputFilename)
	if err != nil {
		log.Fatalf("Reading README failed: %v", err)
	}
	parser := goldmark.DefaultParser()
	root := parser.Parse(text.NewReader(readme))

	// Walk the markdown to identify the (TOML) parts to replace
	blocksToReplace := make([]*includeBlock, 0)
	for rawnode := root.FirstChild(); rawnode != nil; rawnode = rawnode.NextSibling() {
		// Only match TOML code nodes
		var txt []byte
		var start, stop int
		var newlines bool
		var re *regexp.Regexp
		switch node := rawnode.(type) {
		case *ast.FencedCodeBlock:
			if string(node.Language(readme)) != "toml" {
				// Ignore any other node type or language
				continue
			}
			// Extract the block borders
			start = node.Info.Segment.Stop + 1
			stop = start
			lines := node.Lines()
			if lines.Len() > 0 {
				stop = lines.At(lines.Len() - 1).Stop
			}
			txt = node.Info.Text(readme)
			re = tomlIncludesEx
		case *ast.Heading:
			if node.ChildCount() < 2 {
				continue
			}
			child, ok := node.LastChild().(*ast.RawHTML)
			if !ok || child.Segments.Len() == 0 {
				continue
			}
			segment := child.Segments.At(0)
			if !commentIncludesEx.Match(segment.Value(readme)) {
				continue
			}
			start = segment.Stop + 1
			stop = len(readme) // necessary for cases with no more headings
			for rawnode = rawnode.NextSibling(); rawnode != nil; rawnode = rawnode.NextSibling() {
				if h, ok := rawnode.(*ast.Heading); ok && h.Level <= node.Level {
					if rawnode.Lines().Len() > 0 {
						stop = rawnode.Lines().At(0).Start - h.Level - 1
					} else {
						log.Printf("heading without lines: %s", string(rawnode.Text(readme)))
						stop = start // safety measure to prevent removing all text
					}
					break
				}
			}
			txt = segment.Value(readme)
			re = commentIncludesEx
			newlines = true
		default:
			// Ignore everything else
			continue
		}

		// Extract the includes from the node
		block := extractIncludeBlock(txt, re, includeRoot)
		if block != nil {
			block.Start = start
			block.Stop = stop
			block.Newlines = newlines
			blocksToReplace = append(blocksToReplace, block)
		}

		// Catch the case of heading-end-search exhausted all nodes
		if rawnode == nil {
			break
		}
	}

	// Replace the content of the TOML blocks with includes
	var output bytes.Buffer
	output.Grow(len(readme))
	offset := 0
	for _, b := range blocksToReplace {
		// Copy everything up to the beginning of the block we want to replace and make sure we get a newline
		if _, err := output.Write(readme[offset:b.Start]); err != nil {
			log.Fatalf("Writing non-replaced content failed: %v", err)
		}
		if !bytes.HasSuffix(output.Bytes(), []byte("\n")) {
			if _, err := output.Write([]byte("\n")); err != nil {
				log.Fatalf("Writing failed: %v", err)
			}
		}
		offset = b.Stop

		// Insert the include file
		if err := insertIncludes(&output, b); err != nil {
			log.Fatal(err)
		}
	}
	// Copy the remaining of the original file...
	if _, err := output.Write(readme[offset:]); err != nil {
		log.Fatalf("Writing remaining content failed: %v", err)
	}

	// Write output with same permission as input
	file, err := os.OpenFile(inputFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		log.Fatalf("Opening output file failed: %v", err)
	}
	defer file.Close()
	if _, err := output.WriteTo(file); err != nil {
		log.Fatalf("Writing output file failed: %v", err)
	}
}
