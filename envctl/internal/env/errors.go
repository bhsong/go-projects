package env

import "fmt"

type EnvParseError struct {
	Line int
	Msg  string
}

func (e *EnvParseError) Error() string {
	return fmt.Sprintf("줄 %d: %s", e.Line, e.Msg)
}
