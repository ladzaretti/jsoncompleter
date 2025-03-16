# jsoncompleter

[![Go Reference](https://pkg.go.dev/badge/github.com/ladzaretti/jsoncompleter.svg)](https://pkg.go.dev/github.com/ladzaretti/jsoncompleter)[![Go Report Card](https://goreportcard.com/badge/github.com/ladzaretti/jsoncompleter)](https://goreportcard.com/report/github.com/ladzaretti/jsoncompleter)

The `jsoncompleter` repository includes both the `jsoncompleter` package and the `jr` CLI tool, providing a simple way to complete truncated `JSON` strings.

## Motivation

Too many times, I encountered issues with truncated `JSON` entries in `JSONL` (`JSON` Lines) log dumps, making debugging with tools like `jq` and others frustrating.

That's where `jr`—a JSON completer—comes in, completing truncated `JSON` strings to make log analysis easier.

## Installation

### Package `jsoncompleter`
```bash
go get github.com/ladzaretti/jsoncompleter
```

### `jr` cli
```bash
go install github.com/ladzaretti/jsoncompleter/cmd/jr@latest
```