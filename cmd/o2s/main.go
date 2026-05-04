// Package main is the entrypoint for the O2S CLI binary.
package main

import (
	"fmt"
	"os"

	"github.com/O2S-Y/o2s/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
