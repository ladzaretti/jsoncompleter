package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ladzaretti/jsoncompleter"
)

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

	for in.Scan() {
		fmt.Println(completer.Complete(in.Text()))
	}

	if err := in.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

type config struct {
	truncationMarker string
	markTruncation   bool
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

	flag.Usage = usage

	flag.Parse()

	return config
}

func usage() {
	usage := "jr - commandline tool for completing truncated JSON lines.\n\n" +
		"Usage: jr [options] [strings...]\n" +
		"  -m, --mark\t\tEnable marking of truncated JSON lines\n" +
		"  -p, --placeholder\tCustom placeholder for marking truncation\n\n" +
		"Note:\n" +
		"  Assumes the input is a valid but truncated JSON string."
	fmt.Println(usage)
}
