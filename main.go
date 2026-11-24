package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("squish", flag.ContinueOnError)
	fs.SetOutput(errOut)
	sep := fs.String("sep", " ", "separator used to join non-JSON lines")
	raw := fs.Bool("raw", false, "collapse whitespace even when the input is JSON")
	fs.Usage = func() {
		fmt.Fprintf(errOut, `squish turns multi-line text into a single line.

Usage:
  squish [options] [text]

With no text argument, squish reads stdin. A JSON object or array is
compacted. Other text loses shell line-continuations, and whitespace
outside quotes is folded into a single space.

Options:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	var input string
	if fs.NArg() > 0 {
		input = strings.Join(fs.Args(), " ")
	} else {
		b, err := io.ReadAll(in)
		if err != nil {
			fmt.Fprintf(errOut, "squish: read input: %v\n", err)
			return 1
		}
		input = string(b)
	}

	fmt.Fprintln(out, squish(input, *sep, *raw))
	return 0
}
