go-cloudlog
===

[![GoDoc](https://godoc.org/github.com/anexia-it/go-cloudlog?status.svg)](https://godoc.org/github.com/anexia-it/go-cloudlog)
[![Build Status](https://travis-ci.org/anexia-it/go-cloudlog.svg?branch=master)](https://travis-ci.org/anexia-it/go-cloudlog)
[![codecov](https://codecov.io/gh/anexia-it/go-cloudlog/branch/master/graph/badge.svg)](https://codecov.io/gh/anexia-it/go-cloudlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/anexia-it/go-cloudlog)](https://goreportcard.com/report/github.com/anexia-it/go-cloudlog)

go-cloudlog is a client library for Anexia CloudLog.

Currently it only provides to push events to CloudLog. Querying is possible in a future release.

## Install

With a [correctly configured](https://golang.org/doc/install#testing) Go toolchain:

```sh
go get -u github.com/anexia-it/go-cloudlog
```

## Quickstart

```go
package main

import cloudlog "github.com/anexia-it/go-cloudlog"

func main() {

  // Init CloudLog client
  client, err := cloudlog.InitCloudLog("index", "ca.pem", "cert.pem", "cert.key")
  if err != nil {
    panic(err)
  }

  // Push simple message
  client.PushEvent("My first CloudLog event")

  // Push document as map
  logger.PushEvent(map[string]interface{}{
		"timestamp": time.Now(),
		"user":      "test",
		"severity":  1,
		"message":   "My first CloudLog event",
	})

  // Push document as map
  type Document struct {
		Timestamp uint64 `cloudlog:"timestamp"`
		User      string `cloudlog:"user"`
		Severity  int    `cloudlog:"severity"`
		Message   string `cloudlog:"message"`
	}
	logger.PushEvent(&Document{
		Timestamp: 1495171849463,
		User:      "test",
		Severity:  1,
		Message:   "My first CloudLog event",
	})
}
```
