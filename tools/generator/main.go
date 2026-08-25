package main

import (
	"fmt"
	"os"
)

func main() {
	if err := newCommand().Execute(); err != nil {
		if _, printErr := fmt.Fprintln(os.Stderr, "table generation failed:", err); printErr != nil {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
