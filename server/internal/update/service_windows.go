//go:build windows

package update

import (
	"io"
)

func startReplacementProcess(exe string, args []string, stdout, stderr io.Writer) error {
	_, _, _, _ = exe, args, stdout, stderr
	return nil
}
