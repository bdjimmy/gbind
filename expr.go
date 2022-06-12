package gbind

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
)

var (
	dot = []byte{'.'}
)

// NewExecer The function type of the excer generator
type NewExecer func(values [][]byte) (Execer, error)

// Execer A Execer need to implement the methods
type Execer interface {
	Exec(ctx context.Context, value reflect.Value, data interface{}, opt *DefaultOption) (context.Context, error)
	Name() string
}

// ExecerFactory save all of the excer generators
type ExecerFactory struct {
	m map[string]NewExecer
}

// NewExecerFactory save all of the excer generators
func NewExecerFactory() *ExecerFactory {
	return &ExecerFactory{
		m: map[string]NewExecer{},
	}
}

// Regitster your own implemented generator
func (ef *ExecerFactory) Regitster(name string, excerFunc NewExecer) *ExecerFactory {
	ef.m[name] = excerFunc
	return ef
}

// GetExecer get an execer
func (ef *ExecerFactory) GetExecer(value []byte) (execer Execer, err error) {
	var values = bytes.Split(value, dot)
	newExecer, ok := ef.m[SliceToString(values[0])]
	if !ok {
		return nil, fmt.Errorf("syntax error: not support source %s", values[0])
	}
	return newExecer(values)
}
