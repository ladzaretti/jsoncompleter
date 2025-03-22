package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ladzaretti/jsoncompleter"
)

type config struct {
	debug            bool
	skipInvalid      bool
	markTruncation   bool
	truncationMarker string
}

func main() {
	var in *bufio.Scanner
	config := parseConfig()

	args := flag.Args()

	switch {
	case hasPipedStdin():
		in = bufio.NewScanner(os.Stdin)
	case len(args) > 0:
		r := strings.NewReader(strings.Join(args, "\n"))
		in = bufio.NewScanner(r)
	default:
		usage()
		return
	}

	opts := []jsoncompleter.Opt{
		jsoncompleter.WithMarkTruncation(config.markTruncation),
	}
	if len(config.truncationMarker) > 0 {
		opts = append(opts, jsoncompleter.WithTruncationMarker(config.truncationMarker))
	}

	completer := jsoncompleter.New(opts...)

	i := 0

	for in.Scan() {
		i++

		completed := completer.Complete(in.Text())

		if config.skipInvalid && !json.Valid([]byte(completed)) {
			if config.debug {
				fmt.Fprintf(os.Stderr, "%d\n", i)
			}
			continue
		}

		fmt.Println(completed)
	}

	if err := in.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func hasPipedStdin() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func parseConfig() config {
	config := config{}

	flag.BoolVar(&config.markTruncation, "mark", false, "")
	flag.BoolVar(&config.markTruncation, "m", false, "")

	flag.StringVar(&config.truncationMarker, "placeholder", "", "")
	flag.StringVar(&config.truncationMarker, "p", "", "")

	flag.BoolVar(&config.skipInvalid, "skip-invalid", false, "")
	flag.BoolVar(&config.skipInvalid, "s", false, "")

	flag.BoolVar(&config.debug, "debug", false, "")
	flag.BoolVar(&config.debug, "d", false, "")

	flag.Usage = usage

	flag.Parse()

	return config
}

func usage() {
	usage := "jr - commandline tool for completing truncated JSON lines.\n\n" +
		"Usage: jr [options] [strings...]\n" +
		"  -m, --mark\t\tEnable marking of truncated JSON lines\n" +
		"  -p, --placeholder\tSet a custom placeholder for marking truncation\n" +
		"  -s, --skip-invalid\tSkip invalid JSON strings from output\n" +
		"  -d, --debug\t\tPrint the position or line number of skipped invalid JSON strings to stderr\n"
	fmt.Println(usage)
}
