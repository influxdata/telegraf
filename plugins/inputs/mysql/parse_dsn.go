// Copyright 2012 The Go-MySQL-Driver Authors. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package mysql

import (
	"errors"
	"strings"
)

// parseDSN parses the DSN string to a config
func parseDSN(dsn string) (string, error) {
	//var user, passwd string
	var addr, net string

	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	// Find the last '/' (since the password or the net addr might contain a '/')
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '/' {
			var j, k int

			// left part is empty if i <= 0
			if i > 0 {
				// [username[:password]@][protocol[(address)]]
				// Find the last '@' in dsn[:i]
				for j = i; j >= 0; j-- {
					if dsn[j] == '@' {
						// username[:password]
						// Find the first ':' in dsn[:j]
						for k = 0; k < j; k++ {
							if dsn[k] == ':' {
								//passwd = dsn[k+1 : j]
								break
							}
						}
						//user = dsn[:k]

						break
					}
				}

				// [protocol[(address)]]
				// Find the first '(' in dsn[j+1:i]
				for k = j + 1; k < i; k++ {
					if dsn[k] == '(' {
						// dsn[i-1] must be == ')' if an address is specified
						if dsn[i-1] != ')' {
							if strings.ContainsRune(dsn[k+1:i], ')') {
								return "", errors.New("Invalid DSN unescaped")
							}
							return "", errors.New("Invalid DSN Addr")
						}
						addr = dsn[k+1 : i-1]
						break
					}
				}
				net = dsn[j+1 : k]
			}

			break
		}
	}

	// Set default network if empty
	if net == "" {
		net = "tcp"
	}

	// Set default address if empty
	if addr == "" {
		switch net {
		case "tcp":
			addr = "127.0.0.1:3306"
		case "unix":
			addr = "/tmp/mysql.sock"
		default:
			return "", errors.New("Default addr for network '" + net + "' unknown")
		}
	}

	return addr, nil
}
