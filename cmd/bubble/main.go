package main

import (
	"os"

	"bubble/cmd/bubble/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args))
}
