package version

import (
	"fmt"
	"io"
	"runtime/debug"
)

func Print(w io.Writer) {
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Fprintln(w, bi.Main.Version)
		return
	}
	fmt.Fprintln(w, "unknown")
}
