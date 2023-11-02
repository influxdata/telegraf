package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/template"
	"github.com/compose-spec/compose-go/utils"
)

func removeComments(contents []byte) ([]byte, error) {
	tomlReader := bytes.NewReader(contents)

	// Initialize variables for tracking state
	var inQuote, inComment, escaped bool
	var quoteChar byte

	// Initialize buffer for modified TOML data
	var output bytes.Buffer

	buf := make([]byte, 1)
	// Iterate over each character in the file
	for {
		_, err := tomlReader.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		char := buf[0]

		// Toggle the escaped state at backslash to we have true every odd occurrence.
		if char == '\\' {
			escaped = !escaped
		}

		if inComment {
			// If we're currently in a comment, check if this character ends the comment
			if char == '\n' {
				// End of line, comment is finished
				inComment = false
				_, _ = output.WriteRune('\n')
			}
		} else if inQuote {
			// If we're currently in a quote, check if this character ends the quote
			if char == quoteChar && !escaped {
				// End of quote, we're no longer in a quote
				inQuote = false
			}
			output.WriteByte(char)
		} else {
			// Not in a comment or a quote
			if (char == '"' || char == '\'') && !escaped {
				// Start of quote
				inQuote = true
				quoteChar = char
				output.WriteByte(char)
			} else if char == '#' && !escaped {
				// Start of comment
				inComment = true
			} else {
				// Not a comment or a quote, just output the character
				output.WriteByte(char)
			}
		}

		// Reset escaping if any other character occurred
		if char != '\\' {
			escaped = false
		}
	}
	return output.Bytes(), nil
}

func substituteEnvironment(contents []byte, oldReplacementBehavior bool) ([]byte, error) {
	options := []template.Option{
		template.WithReplacementFunction(func(s string, m template.Mapping, cfg *template.Config) (string, error) {
			result, applied, err := template.DefaultReplacementAppliedFunc(s, m, cfg)
			if err == nil && !applied {
				// Keep undeclared environment-variable patterns to reproduce
				// pre-v1.27 behavior
				return s, nil
			}
			if err != nil && strings.HasPrefix(err.Error(), "Invalid template:") {
				// Keep invalid template patterns to ignore regexp substitutions
				// like ${1}
				return s, nil
			}
			return result, err
		}),
		template.WithoutLogging,
	}
	if oldReplacementBehavior {
		options = append(options, template.WithPattern(oldVarRe))
	}

	envMap := utils.GetAsEqualsMap(os.Environ())
	retVal, err := template.SubstituteWithOptions(string(contents), func(k string) (string, bool) {
		if v, ok := envMap[k]; ok {
			return v, ok
		}
		return "", false
	}, options...)
	return []byte(retVal), err
}
