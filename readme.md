# jsoncompleter

[![Go Reference](https://pkg.go.dev/badge/github.com/ladzaretti/jsoncompleter.svg)](https://pkg.go.dev/github.com/ladzaretti/jsoncompleter)[![Go Report Card](https://goreportcard.com/badge/github.com/ladzaretti/jsoncompleter)](https://goreportcard.com/report/github.com/ladzaretti/jsoncompleter)

`jsoncompleter` is a Go package that provides a simple way to restore truncated JSON strings.

## Motivation

Too many times, I encountered issues with truncated `JSON` lines in logging agent `JSONL` log dumps, making debugging with tools like `jq` and others frustrating.

That's where `jlete` comes in - restoring truncated `JSON` strings to make log analysis easier and more efficient.

## Installation

### Package `jsoncompleter`
```bash
go get github.com/ladzaretti/jsoncompleter
```

### `jlete` cli tool
```bash
go install github.com/ladzaretti/jsoncompleter/cmd/jlete@latest
```