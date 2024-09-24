package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/url"
	"os"
)

var decode bool

func main() {
	flag.BoolVar(&decode, "d", false, "Decode the input")
	flag.Parse()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(os.Stdin); err != nil {
		fmt.Printf("Error reading from stdin: %v\n", err)
		os.Exit(1)
	}

	if decode {
		out, err := url.QueryUnescape(buf.String())
		if err != nil {
			fmt.Printf("Error decoding: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}

	fmt.Println(url.QueryEscape(buf.String()))
}
