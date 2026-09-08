package main

import (
	"fmt"
	"os"

	"github.com/arran4/lookup/cmd/internal/cli"
)

func main() {
	if err := cli.Run("json-simple-path", os.Args[1:], os.Stdin, os.Stdout, os.Stderr, "JSON"); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
