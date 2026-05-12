package main

import (
	"os"

	"github.com/Xusk947/bubble/cmd/bubble/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args))
}
