package cli

import "errors"

const (
	ExitRuntime            = 1
	ExitUsage              = 2
	ExitIO                 = 3
	ExitMissingCredentials = 4
)

type codedError struct {
	code int
	err  error
}

func (e codedError) Error() string { return e.err.Error() }
func (e codedError) Unwrap() error { return e.err }

func withCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return codedError{code: code, err: err}
}

func ExitCode(err error) int {
	var coded codedError
	if errors.As(err, &coded) {
		return coded.code
	}
	return ExitRuntime
}
