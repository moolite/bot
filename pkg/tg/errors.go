package tg

import "fmt"

type ErrNoMethod struct {
	fn string
}

func (e *ErrNoMethod) Error() string {
	return fmt.Sprintf("error in sendable: no method defined. fn: %s", e.fn)
}
