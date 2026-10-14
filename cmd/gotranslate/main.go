package main

import (
	"fmt"
	"os"

	"github.com/shepard-labs/go-translation/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		code := cli.ExitCode(err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}
