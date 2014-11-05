package main

import (
	"flag"
	"os"
	"fmt"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [dir]\n", os.Args[0])
}

func buildFeed() {

}

func main() {
	flag.Parse()

	if (flag.NArg() > 1) {
		usage()
		os.Exit(-1)
	}

	var workDir = os.Args[0]
	if (flag.NArg() == 1) {
		workDir = flag.Arg(0)
	}

	fmt.Println("WorkDir ", workDir)
}
