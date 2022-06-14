package gbind

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type exprKey struct{}

func TestExpr(t *testing.T) {
	ef := newexecerFactory().regitster("simple", NewSimpleExecer)
	excer, err := ef.getExecer([]byte("simple.key"))
	assert.Nil(t, err)

	var v string
	excer.Exec(context.WithValue(context.Background(), exprKey{}, "simple-k-d"), reflect.ValueOf(&v).Elem(), nil, nil)
	assert.Equal(t, "simple-k-d", v)

}

func NewSimpleExecer(values [][]byte) (Execer, error) {
	n := len(values)
	if n != 2 {
		return nil, errors.New("syntax error: simple error")
	}
	switch {
	case bytes.Equal(values[1], []byte("key")):
		return &simpleKeyExecer{}, nil
	}
	return nil, fmt.Errorf("syntax error: not support simple %s", values[1])
}

type simpleKeyExecer struct{}

// Exec
func (s *simpleKeyExecer) Exec(ctx context.Context, value reflect.Value, data interface{}, opt *DefaultOption) (context.Context, error) {
	err := TrySet(value, []string{ctx.Value(exprKey{}).(string)}, opt)
	return ctx, err
}

// Name
func (s *simpleKeyExecer) Name() string {
	return "simple.key"
}
