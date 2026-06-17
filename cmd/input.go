package cmd

import (
	"fmt"
	"io"
	"os"
)

// readTextArg resolves a text flag value that may reference external input:
//
//	"-"      read all of stdin
//	"@path"  read the file at path
//	other    used verbatim
//
// To pass a literal value beginning with "@", read it from stdin or a file.
func readTextArg(val string, stdin io.Reader) (string, error) {
	switch {
	case val == "-":
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(data), nil
	case len(val) > 1 && val[0] == '@':
		path := val[1:]
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", path, err)
		}
		return string(data), nil
	default:
		return val, nil
	}
}
