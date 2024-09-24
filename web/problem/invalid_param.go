package problem

import (
	"fmt"
)

var (
	_ error = (*InvalidParam)(nil)
)

// InvalidParam represents a field error
type InvalidParam struct {
	Err   error  `json:"-"`
	Value string `json:"-"`

	Name   string `json:"name,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func (e InvalidParam) Error() string {
	name, reason := e.Name, e.Reason
	if name == "" {
		name = "undefined"
	}
	if reason == "" {
		reason = "undefined"
	}

	return fmt.Sprintf("%s: %s", name, reason)
}

func (e InvalidParam) Unwrap() error { return e.Err }

// NewInvalidParamError ...
func NewInvalidParamError(err error, name, value, reason string) *InvalidParam {
	return &InvalidParam{
		Err:    err,
		Name:   name,
		Value:  value,
		Reason: reason,
	}
}
