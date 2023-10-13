go-cloudlog
===

[![Go Reference](https://pkg.go.dev/badge/anexia-it/go-cloudlog?status.svg)](https://pkg.go.dev/github.com/anexia-it/go-cloudlog)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/anexia-it/go-cloudlog)](https://github.com/anexia-it/go-cloudlog/releases/latest)
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/anexia-it/go-cloudlog)](https://github.com/anexia-it/go-cloudlog/blob/master/go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/anexia-it/go-cloudlog)](https://goreportcard.com/report/github.com/anexia-it/go-cloudlog)
[![GitHub license](https://img.shields.io/github/license/anexia-it/go-cloudlog)](https://github.com/anexia-it/go-cloudlog/blob/master/LICENSE)

go-cloudlog is a client library for Anexia CloudLog.

## Install

With a [correctly configured](https://golang.org/doc/install#testing) Go toolchain:

```sh
go get -u github.com/anexia-it/go-cloudlog/v2
```

## Quickstart

```go
package main

import (
	"github.com/anexia-it/go-cloudlog/v2"
	"time"
)

func main() {
	// Init CloudLog client
	client, err := cloudlog.NewCloudLog("index", "token")
	if err != nil {
		panic(err)
	}

	// Push simple message
	client.PushEvent("My first CloudLog event")

	// Push document as map
	client.PushEvent(map[string]interface{}{
		"timestamp": time.Now(),
		"user":      "test",
		"severity":  1,
		"message":   "My CloudLog event with a map",
	})

	// Push document as struct
	type Document struct {
		Timestamp uint64 `cloudlog:"timestamp"`
		User      string `cloudlog:"user"`
		Severity  int    `cloudlog:"severity"`
		Message   string `cloudlog:"message"`
	}
	client.PushEvent(&Document{
		Timestamp: 1495171849463,
		User:      "test",
		Severity:  1,
		Message:   "My CloudLog event with a struct",
	})
}
```
