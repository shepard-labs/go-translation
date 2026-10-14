package input

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

type Options struct {
	Args  []string
	Text  string
	File  string
	Stdin io.Reader
}

func Read(opts Options) (string, error) {
	sources := 0
	if len(opts.Args) > 0 {
		sources++
	}
	if opts.Text != "" {
		sources++
	}
	if opts.File != "" {
		sources++
	}
	stdinText := ""
	if opts.Stdin != nil && isPipedStdin() {
		data, err := io.ReadAll(opts.Stdin)
		if err != nil {
			return "", err
		}
		stdinText = string(data)
		if stdinText != "" {
			sources++
		}
	}

	if sources == 0 {
		return "", errors.New("missing input: pass text, --text, --file, or pipe stdin")
	}
	if sources > 1 {
		return "", errors.New("input conflict: use only one of positional text, --text, --file, or stdin")
	}

	if len(opts.Args) > 0 {
		return opts.Args[0], nil
	}
	if opts.Text != "" {
		return opts.Text, nil
	}
	if opts.File != "" {
		data, err := os.ReadFile(opts.File)
		if err != nil {
			return "", err
		}
		if looksBinary(data) {
			return "", fmt.Errorf("input file appears to be binary: %s", opts.File)
		}
		return string(data), nil
	}
	return stdinText, nil
}

func isPipedStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}

func looksBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if !utf8.Valid(data) {
		return true
	}
	limit := len(data)
	if limit > 8000 {
		limit = 8000
	}
	for _, b := range data[:limit] {
		if b == 0 {
			return true
		}
	}
	return false
}
