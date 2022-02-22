package xerrors

import (
	"fmt"

	"github.com/tableauio/loader/_lab/go/tableau/code"
)

type Error struct {
	code  code.Code
	cause error
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, cause: %+v", e.code, e.cause)
}

func Errorf(code code.Code, format string, args ...interface{}) error {
	return &Error{
		code:  code,
		cause: fmt.Errorf(format, args...),
	}
}

func Code(err error) code.Code {
	if ferr, ok := err.(*Error); ok {
		return ferr.code
	}
	return code.Unknown
}

func Is(err error, code code.Code) bool {
	return Code(err) == code
}
