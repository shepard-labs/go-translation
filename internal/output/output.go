package output

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func Write(text, path string, force bool, stdout io.Writer) error {
	if path == "" {
		_, err := fmt.Fprint(stdout, text)
		return err
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("output file exists: pass --force to overwrite %s", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return os.WriteFile(path, []byte(text), 0o644)
}
