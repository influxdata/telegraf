// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package like2regexp

import (
	"strings"
)

const special = `\.+*?()|{}^$` // All regexp special chars excetp `[]`

func WMILikeToRegexp(like string) string {
	var buf strings.Builder
	// Quote special characters
	inclass := false
	for i := 0; i < len(like); i++ {
		c := like[i]

		if inclass && c == ']' {
			inclass = false
		} else if c == '[' {
			inclass = true
		} else if !inclass {
			switch c {
			case '_':
				c = '.'
			case '%':
				buf.WriteByte('.')
				c = '*'
			default:
				if strings.IndexByte(special, c) != -1 { // Escape special chars
					buf.WriteByte('\\')
				}
			}
		}
		buf.WriteByte(c)
	}

	return `(?i:^` + buf.String() + `$)`
}