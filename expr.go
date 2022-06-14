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

// execerFactory save all of the excer generators
type execerFactory struct {
	m map[string]NewExecer
}

// newexecerFactory save all of the excer generators
func newexecerFactory() *execerFactory {
	return &execerFactory{
		m: map[string]NewExecer{},
	}
}

// regitster your own implemented generator
func (ef *execerFactory) regitster(name string, excerFunc NewExecer) *execerFactory {
	ef.m[name] = excerFunc
	return ef
}

// getExecer get an execer
func (ef *execerFactory) getExecer(value []byte) (execer Execer, err error) {
	var values = bytes.Split(value, dot)
	newExecer, ok := ef.m[SliceToString(values[0])]
	if !ok {
		return nil, fmt.Errorf("syntax error: not support source %s", values[0])
	}
	return newExecer(values)
}
