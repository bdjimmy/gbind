package gbind

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ---------- http execer start ----------
var (
	_ Execer = &httpPathExcer{}
	_ Execer = &httpQueryExcer{}
	_ Execer = &httpCookieExcer{}
	_ Execer = &httpFormExcer{}
	_ Execer = &httpHeadExcer{}
)

var (
	httpPathID   = []byte("path")   // http.path
	httpHeadID   = []byte("header") // http.head.Refer
	httpCookieID = []byte("cookie") // http.cookie.BDUSS
	httpPostID   = []byte("form")   // http.form.zid
	httpQueryID  = []byte("query")  // http.query.appid

	errHTTP       = errors.New("syntax error: http error")
	errHTTPQuery  = errors.New("syntax error: http query error")
	errHTTPPath   = errors.New("syntax error: http path error")
	errHTTPHead   = errors.New("syntax error: http head error")
	errHTTPCookie = errors.New("syntax error: http cookie error")
	errHTTPPost   = errors.New("syntax error: http post error")
)

// newHTTPExecer Execer are generated based on the values
func newHTTPExecer(values [][]byte) (Execer, error) {
	n := len(values)
	if n < 2 {
		return nil, errHTTP
	}
	switch {
	case bytes.Equal(values[1], httpQueryID):
		if n != 3 {
			return nil, errHTTPQuery
		}
		return &httpQueryExcer{
			param: SliceToString(values[2]),
		}, nil
	case bytes.Equal(values[1], httpPathID):
		if n != 2 {
			return nil, errHTTPPath
		}
		return &httpPathExcer{}, nil
	case bytes.Equal(values[1], httpHeadID):
		if n != 3 {
			return nil, errHTTPHead
		}
		return &httpHeadExcer{
			param: SliceToString(values[2]),
		}, nil
	case bytes.Equal(values[1], httpCookieID):
		if n != 3 {
			return nil, errHTTPCookie
		}
		return &httpCookieExcer{
			param: SliceToString(values[2]),
		}, nil
	case bytes.Equal(values[1], httpPostID):
		if n != 3 {
			return nil, errHTTPPost
		}
		return &httpFormExcer{
			param: SliceToString(values[2]),
		}, nil
	}
	return nil, fmt.Errorf("syntax error: not support http %s", values[1])
}

// ----------------- http.path -----------------
type httpPathExcer struct{}

// Exec
func (h *httpPathExcer) Exec(ctx context.Context, value reflect.Value, data interface{}, opt *DefaultOption) (context.Context, error) {
	req, ok := data.(*http.Request)
	if !ok {
		return ctx, errHTTPPath
	}
	err := TrySet(value, []string{req.URL.Path}, opt)
	return ctx, err
}

// Name
func (h *httpPathExcer) Name() string {
	return "http.path"
}

// ----------------- http.query -----------------
type httpQueryExcer struct {
	param string
}

func (h *httpQueryExcer) Exec(ctx context.Context, value reflect.Value, data interface{}, opt *DefaultOption) (context.Context, error) {
	req, ok := data.(*http.Request)
	if !ok {
		return ctx, errors.New("data is not a pointer of http.Request")
	}
	ctx = newHTTPContext(ctx, req)
	vs := mustContextHTTPMeta(ctx).getQueryArray(h.param)
	err := TrySet(value, vs, opt)
	return ctx, err
}

func (h *httpQueryExcer) Name() string {
	return "http.query"
}

// ----------------- http.head -----------------
type httpHeadExcer struct {
	param string
}

func (h *httpHeadExcer) Exec(ctx context.Context, value reflect.Value, data interface{}, opt *DefaultOption) (context.Context, error) {
	req, ok := data.(*http.Request)
	if !ok {
		return ctx, errors.New("data is not a pointer of http.Request")
	}
	return ctx, TrySet(value, req.Header.Values(h.param), opt)
}

func (h *httpHeadExcer) Name() string {
	return "http.head"
}

// ----------------- http.form -----------------
type httpFormExcer struct {
	param string
}

func (h *httpFormExcer) Exec(ctx context.Context, value reflect.Value, data interface{}, opt *DefaultOption) (context.Context, error) {
	req, ok := data.(*http.Request)
	if !ok {
		return ctx, errors.New("data is not a pointer of http.Request")
	}
	ctx = newHTTPContext(ctx, req)
	vs := mustContextHTTPMeta(ctx).getFormArray(h.param)
	err := TrySet(value, vs, opt)
	return ctx, err
}

func (h *httpFormExcer) Name() string {
	return "http.form"
}

// ----------------- http.cookie -----------------
type httpCookieExcer struct {
	param string
}

func (h *httpCookieExcer) Exec(ctx context.Context, value reflect.Value, data interface{}, opt *DefaultOption) (context.Context, error) {
	req, ok := data.(*http.Request)
	if !ok {
		return ctx, errors.New("data is not a pointer of http.Request")
	}
	if c, err := req.Cookie(h.param); err == nil {
		v, _ := url.QueryUnescape(c.Value)
		err := TrySet(value, []string{v}, opt)
		return ctx, err
	}
	err := TrySet(value, []string{}, opt)
	return ctx, err

}

func (h *httpCookieExcer) Name() string {
	return "http.cookie"
}

// --------- http execer end ---------

// DefaultOption options for the default values
type DefaultOption struct {
	IsDefaultExists  bool
	DefaultValue     string
	DefaultSplitFlag string
}

// TrySet try to set up the value
// A custom callback function can invoke this function
func TrySet(value reflect.Value, vs []string, opt *DefaultOption) error {
	var def = false
	if opt != nil && opt.IsDefaultExists {
		def = true
	}
	if len(vs) == 0 && !def {
		return nil
	}

	if len(vs) == 0 && def {
		vs = strings.Split(opt.DefaultValue, opt.DefaultSplitFlag)
	}
	switch value.Interface().(type) {
	case time.Duration:
		return setTimeDuration(vs, value)
	}

	switch value.Kind() {
	case reflect.Slice:
		return setSlice(vs, value)
	case reflect.Array:
		if len(vs) != value.Len() {
			return fmt.Errorf("%q is not valid value for %s", vs, value.Type().String())
		}
		return setArray(vs, value)
	default:
		var val string
		if len(vs) > 0 {
			val = vs[0]
		}
		return setWithProperType(val, value)
	}
}

func setWithProperType(val string, value reflect.Value) error {
	switch value.Kind() {
	case reflect.Bool:
		return setBoolField(val, value)
	// float
	case reflect.Float32:
		return setFloatField(val, 32, value)
	case reflect.Float64:
		return setFloatField(val, 64, value)
	// int
	case reflect.Int:
		return setIntField(val, 0, value)
	case reflect.Int8:
		return setIntField(val, 8, value)
	case reflect.Int16:
		return setIntField(val, 16, value)
	case reflect.Int32:
		return setIntField(val, 32, value)
	case reflect.Int64:
		return setIntField(val, 64, value)
	// uint
	case reflect.Uint:
		return setUintField(val, 0, value)
	case reflect.Uint8:
		return setUintField(val, 8, value)
	case reflect.Uint16:
		return setUintField(val, 16, value)
	case reflect.Uint32:
		return setUintField(val, 32, value)
	case reflect.Uint64:
		return setUintField(val, 64, value)
	// string
	case reflect.String:
		value.SetString(val)
	}
	return nil
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setTimeDuration(vals []string, field reflect.Value) error {
	if len(vals) != 1 {
		return nil
	}
	d, err := time.ParseDuration(vals[0])
	if err == nil {
		field.Set(reflect.ValueOf(d))
	}
	return err
}

func setArray(vals []string, value reflect.Value) error {
	for i, s := range vals {
		err := setWithProperType(s, value.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func setSlice(vals []string, value reflect.Value) error {
	slice := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := setArray(vals, slice)
	if err != nil {
		return err
	}
	value.Set(slice)
	return nil
}
