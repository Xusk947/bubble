package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/Xusk947/bubble/cmd/bubble/internal/atlas"
	"github.com/Xusk947/bubble/cmd/bubble/internal/initcmd"
	"github.com/Xusk947/bubble/cmd/bubble/internal/usage"
	"github.com/Xusk947/bubble/cmd/bubble/internal/version"
)

type exitCode int

const (
	exitOK      exitCode = 0
	exitUsage   exitCode = 2
	exitRuntime exitCode = 1
)

func Run(argv []string) int {
	return int(run(argv, os.Stdout, os.Stderr))
}

func run(argv []string, stdout io.Writer, stderr io.Writer) exitCode {
	if len(argv) < 2 {
		usage.Print(stderr)
		return exitUsage
	}

	switch argv[1] {
	case "version":
		version.Print(stdout)
		return exitOK
	case "init":
		if err := initcmd.Run(argv[2:], stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return exitRuntime
		}
		return exitOK
	case "db":
		if err := atlas.Run(argv[2:], stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return exitRuntime
		}
		return exitOK
	case "atlas":
		if err := atlas.Run(argv[2:], stdout, stderr); err != nil {
			fmt.Fprintln(stderr, err)
			return exitRuntime
		}
		return exitOK
	case "help", "-h", "--help":
		usage.Print(stdout)
		return exitOK
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", argv[1])
		usage.Print(stderr)
		return exitUsage
	}
}
