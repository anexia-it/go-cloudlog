go-cloudlog
===

[![GoDoc](https://godoc.org/github.com/anexia-it/go-cloudlog?status.svg)](https://godoc.org/github.com/anexia-it/go-cloudlog)
[![Build Status](https://travis-ci.org/anexia-it/go-cloudlog.svg?branch=master)](https://travis-ci.org/anexia-it/go-cloudlog)
[![codecov](https://codecov.io/gh/anexia-it/go-go-cloudlog/branch/master/graph/badge.svg)](https://codecov.io/gh/anexia-it/go-cloudlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/anexia-it/go-cloudlog)](https://goreportcard.com/report/github.com/anexia-it/go-cloudlog)

go-cloudlog is a client library for Anexia CloudLog.

Currently it only provides to push events to CloudLog. Querying is possible in a future release.

## Install

With a [correctly configured](https://golang.org/doc/install#testing) Go toolchain:

```sh
go get -u github.com/anexia-it/go-cloudlog
```

## Examples

You can find examples located in the `examples` directory

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
  client.PushEvent("message")
  
  // Push document
  client.PushEvent(`{
    "timestamp": 1495024205123,
    "user": "test",
    "severity": 1,
    "message": "My first CloudLog event"
  }`)
}
```
