# TeamSpeak 3 [![Go Report Card](https://goreportcard.com/badge/github.com/multiplay/go-ts3)](https://goreportcard.com/report/github.com/multiplay/go-ts3) [![License](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/multiplay/go-ts3/blob/master/LICENSE) [![GoDoc](https://godoc.org/github.com/multiplay/go-ts3?status.svg)](https://godoc.org/github.com/multiplay/go-ts3) [![Build Status](https://travis-ci.org/multiplay/go-ts3.svg?branch=master)](https://travis-ci.org/multiplay/go-ts3)

go-ts3 is a [Go](http://golang.org/) client for the [TeamSpeak 3 ServerQuery Protocol](http://media.teamspeak.com/ts3_literature/TeamSpeak%203%20Server%20Query%20Manual.pdf).

Features
--------
* [ServerQuery](http://media.teamspeak.com/ts3_literature/TeamSpeak%203%20Server%20Query%20Manual.pdf) Support.

Installation
------------
```sh
go get -u github.com/multiplay/go-ts3
```

Examples
--------

Using go-ts3 is simple just create a client, login and then send commands e.g.
```go
package main

import (
	"log"

        "github.com/multiplay/go-ts3"
)

func main() {
        c, err := ts3.NewClient("192.168.1.102:10011")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := c.Login(user, pass); err != nil {
		log.Fatal(err)
	}

	if v, err := c.Version(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("server is running:", v)
	}
}
```

Documentation
-------------
- [GoDoc API Reference](http://godoc.org/github.com/multiplay/go-ts3).

License
-------
go-ts3 is available under the [BSD 2-Clause License](https://opensource.org/licenses/BSD-2-Clause).
```
