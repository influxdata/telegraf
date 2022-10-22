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
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type includeBlock struct {
	Includes []string
	Start    int
	Stop     int
}

func (b *includeBlock) extractBlockBorders(node *ast.FencedCodeBlock) {
	// The node info starts at the language tag and stops right behind it
	b.Start = node.Info.Segment.Stop + 1
	b.Stop = b.Start

	// To determine the end of the block, we need to iterate to the last line
	// and take its stop-offset as the end of the block.
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		b.Stop = lines.At(i).Stop
	}
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

func insertIncludes(buf *bytes.Buffer, b includeBlock) error {
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

	return nil
}

func main() {
	// Finds all TOML sections of the form `toml @includefile` and extracts the `includefile` part
	tomlIncludesEx := regexp.MustCompile(`^toml\s+(@.+)+$`)
	tomlIncludeMatch := regexp.MustCompile(`(?:@([^\s]+))+`)

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
	blocksToReplace := make([]includeBlock, 0)
	for node := root.FirstChild(); node != nil; node = node.NextSibling() {
		// Only match TOML code nodes
		codeNode, ok := node.(*ast.FencedCodeBlock)
		if !ok || string(codeNode.Language(readme)) != "toml" {
			// Ignore any other node type or language
			continue
		}

		// Extract the includes from the node
		includes := tomlIncludesEx.FindSubmatch(codeNode.Info.Text(readme))
		if len(includes) != 2 {
			continue
		}
		block := includeBlock{}
		for _, inc := range tomlIncludeMatch.FindAllSubmatch(includes[1], -1) {
			if len(inc) != 2 {
				continue
			}
			include := string(inc[1])
			// Safeguards to avoid directory traversals and other bad things
			if strings.ContainsRune(include, os.PathSeparator) {
				log.Printf("Ignoring include %q for containing a path...", include)
				continue
			}
			if fi, err := os.Stat(include); err != nil || !fi.Mode().IsRegular() {
				log.Printf("Ignoring include %q as it cannot be found or is not a regular file...", include)
				continue
			}
			block.Includes = append(block.Includes, string(inc[1]))
		}

		// Extract the block borders
		block.extractBlockBorders(codeNode)
		blocksToReplace = append(blocksToReplace, block)
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
