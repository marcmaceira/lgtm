package main

import (
	"os"

	"lgtm/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
