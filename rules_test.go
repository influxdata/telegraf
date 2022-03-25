package main

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
	"github.com/stretchr/testify/require"
)

func TestMainHeading(t *testing.T) {
	md := []byte("## markdown document")
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
	root := p.Parse(md)
	tester := T{
		filename: "filename",
		ruleName: "rulename",
	}
	require.NoError(t, mainHeading(&tester, root))
}
