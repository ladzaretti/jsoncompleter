# jsoncompleter

[![Go Reference](https://pkg.go.dev/badge/github.com/ladzaretti/jsoncompleter.svg)](https://pkg.go.dev/github.com/ladzaretti/jsoncompleter)[![Go Report Card](https://goreportcard.com/badge/github.com/ladzaretti/jsoncompleter)](https://goreportcard.com/report/github.com/ladzaretti/jsoncompleter)

The `jsoncompleter` repository includes both the `jsoncompleter` package and the `jr` CLI tool, providing a simple way to complete truncated `JSON` strings.

[Read more on pkg.go.dev.](https://pkg.go.dev/github.com/ladzaretti/jsoncompleter)

## Motivation

Too many times, I encountered issues with truncated `JSON` entries in `JSONL` (`JSON` Lines) log dumps, making debugging with tools like `jq` and others frustrating.

That's where `jr`—a JSON completer—comes in, completing truncated `JSON` strings to make log analysis easier.

### Benefits of `jr`

Truncated `JSON` data might still contain useful information, `jr` ensures that no data is discarded due to truncation. 

It marks missing parts with placeholders, making it easy to track what's missing.

Once completed, processing continues as expected, without skipping valuable insights from your data.

## Installation

### Package `jsoncompleter`
```bash
go get github.com/ladzaretti/jsoncompleter
```

### The `jr` CLI

#### Using Go
```bash
go install github.com/ladzaretti/jsoncompleter/cmd/jr@latest
```

#### Binary releases
Precompiled binaries for various architectures are available on the release page.
- https://github.com/ladzaretti/jsoncompleter/releases


### Usage
```bash
$ jr -h
jr - commandline tool for completing truncated JSON lines.

Usage: jr [options] [strings...]
  -m, --mark            Enable marking of truncated JSON lines
  -p, --placeholder     Set a custom placeholder for marking truncation
  -s, --skip-invalid    Skip invalid JSON strings from output
  -d, --debug           Print the position or line number of skipped invalid JSON strings to stderr

Note:
  Assumes the input is a valid but truncated JSON string.
```


## Example:

Without `jr`

This demonstrates the error when parsing a truncated JSON string, where data after the error is not processed.

This JSONL input has 3 lines: the first and last are valid, the second is truncated.

```bash
$ echo -e '{"foo":"bar"}\n{"baz":\n{"qux":null}' | jq -c
{"foo":"bar"}
jq: parse error: Unfinished JSON term at EOF at line 4, column 0
```

With `jr`

This will complete the truncated JSON and allow the parsing to proceed.

```bash
$ echo -e '{"foo":"bar"}\n{"baz":\n{"qux":null}' | jr -m | jq -c
{"foo":"bar"}
{"baz":"","__TRUNCATION_MARKER__":""}
{"qux":null}
```