package main

import (
	"fmt"
	"os"

	"github.com/arran4/lookup/cmd/internal/cli"
)

func main() {
	if err := cli.Run("yaml-simpe-path", os.Args[1:], os.Stdin, os.Stdout, os.Stderr, "YAML"); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
