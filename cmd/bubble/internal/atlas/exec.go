package atlas

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func execAtlas(ctx context.Context, bin string, args []string, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("atlas failed: %w", err)
		}
		return fmt.Errorf("atlas exec failed: %w", err)
	}
	return nil
}

