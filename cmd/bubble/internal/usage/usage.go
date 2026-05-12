package usage

import (
	"fmt"
	"io"
)

func Print(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  bubble version")
	fmt.Fprintln(w, "  bubble init [--module <module>] [--bubble-module <module>] [--dir <path>] [--name <app_name>] [--db <none|sqlite|postgres>] [--cache <local|redis>] [--queue <none|nats|kafka>] [--s3] [--skip-go] [--non-interactive]")
	fmt.Fprintln(w, "  bubble db diff --name <name> --dev-url <url> [--to <ent://...>] [--dir <path>] [--dotenv <path>] [--atlas-bin <path>]")
	fmt.Fprintln(w, "  bubble db apply [--url <url>] [--dir <path>] [--dotenv <path>] [--atlas-bin <path>]")
	fmt.Fprintln(w, "  bubble atlas diff ... (alias for bubble db diff)")
	fmt.Fprintln(w, "  bubble atlas apply ... (alias for bubble db apply)")
}
