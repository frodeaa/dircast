package main

import (
	"flag"
	"os"
	"fmt"
	"encoding/xml"
	"path/filepath"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [dir]\n", os.Args[0])
}

type Item struct {
	XMLName xml.Name `xml:"item"`
	Title   string `xml:"title"`
}

func visitFiles() filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {

		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil
		}

		if !!f.IsDir() {
			return nil // not a file. ignore
		}

		matched, err := filepath.Match("*.mp3", f.Name())
		if err != nil {
			fmt.Println(err) // malformed pattern
			return err
		}

		if matched {
			v := &Item{Title: f.Name()}
			output, err := xml.MarshalIndent(v, "  ", "    ")
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}
			os.Stdout.Write(output)
		}
		return nil

	}
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

	err := filepath.Walk(workDir, visitFiles())

	fmt.Println("WorkDir %s, result %v", workDir, err)
}
