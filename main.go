package main

import (
	"os"

	"github.com/arisu-archive/protocol-encoder-go/internal/cli/root"
)

var Version = "0.0.1"

func main() {
	root.Execute(Version, os.Exit, os.Stdin, os.Stdout, os.Stderr, os.Args[1:])
}
