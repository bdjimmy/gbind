package gbind

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHttpCookie(t *testing.T) {
	_, err := newHTTPExecer(bytes.Split([]byte("http.cookie.id.a"), dot))
	assert.NotNil(t, err)

	excer, err := newHTTPExecer(bytes.Split([]byte("http.cookie.id"), dot))
	assert.Nil(t, err)

	for testName, st := range map[string]struct {
		context.Context
		value  interface{}
		req    *http.Request
		opt    *DefaultOption
		expect interface{}
	}{
		"http-cookie-id": {
			context.Background(),
			struct{ T int }{},
			newReq().addCookie("id", "123").r(),
			nil,
			int(123),
		},
		"http-cookie-id-default": {
			context.Background(),
			struct{ T int }{},
			newReq().r(),
			&DefaultOption{IsDefaultExists: true, DefaultValue: "123456", DefaultSplitFlag: "|"},
			int(123456),
		},
	} {
		v := reflect.New(reflect.TypeOf(st.value)).Elem().Field(0)
		_, err := excer.Exec(st.Context, v, st.req, st.opt)
		assert.Nil(t, err, testName)
		assert.Equal(t, st.expect, v.Interface(), testName)
	}
}

func TestHttpHead(t *testing.T) {
	_, err := newHTTPExecer(bytes.Split([]byte("http.header.id.a"), dot))
	assert.NotNil(t, err)

	excer, err := newHTTPExecer(bytes.Split([]byte("http.header.id"), dot))
	assert.Nil(t, err)

	for testName, st := range map[string]struct {
		context.Context
		value  interface{}
		req    *http.Request
		opt    *DefaultOption
		expect interface{}
	}{
		"http-head-id": {
			context.Background(),
			struct{ T int }{},
			newReq().addHeader("id", "123").r(),
			nil,
			int(123),
		},
		"http-head-id-default": {
			context.Background(),
			struct{ T int }{},
			newReq().r(),
			&DefaultOption{IsDefaultExists: true, DefaultValue: "123456", DefaultSplitFlag: "|"},
			int(123456),
		},
	} {
		v := reflect.New(reflect.TypeOf(st.value)).Elem().Field(0)
		_, err := excer.Exec(st.Context, v, st.req, st.opt)
		assert.Nil(t, err, testName)
		assert.Equal(t, st.expect, v.Interface(), testName)
	}
}

func TestHttpPath(t *testing.T) {
	_, err := newHTTPExecer(bytes.Split([]byte("http.path.a"), dot))
	assert.NotNil(t, err)

	excer, err := newHTTPExecer(bytes.Split([]byte("http.path"), dot))
	assert.Nil(t, err)

	for testName, st := range map[string]struct {
		context.Context
		value  interface{}
		req    *http.Request
		opt    *DefaultOption
		expect interface{}
	}{
		"http-path": {
			context.Background(),
			struct{ T string }{},
			newReq().setPath("/api/test").r(),
			nil,
			"/api/test",
		},
	} {
		v := reflect.New(reflect.TypeOf(st.value)).Elem().Field(0)
		_, err := excer.Exec(st.Context, v, st.req, st.opt)
		assert.Nil(t, err, testName)
		assert.Equal(t, st.expect, v.Interface(), testName)
	}
}

func TestHttpForm(t *testing.T) {
	_, err := newHTTPExecer(bytes.Split([]byte("http.form.id.a"), dot))
	assert.NotNil(t, err)

	excer, err := newHTTPExecer(bytes.Split([]byte("http.form.id"), dot))
	assert.Nil(t, err)

	for testName, st := range map[string]struct {
		context.Context
		value  interface{}
		req    *http.Request
		opt    *DefaultOption
		expect interface{}
	}{
		"http-form-id": {
			context.Background(),
			struct{ T int }{},
			newReq().addFormParam("id", "123").r(),
			nil,
			int(123),
		},
		"http-form-id-default": {
			context.Background(),
			struct{ T int }{},
			newReq().r(),
			&DefaultOption{IsDefaultExists: true, DefaultValue: "123456", DefaultSplitFlag: "|"},
			int(123456),
		},
	} {
		v := reflect.New(reflect.TypeOf(st.value)).Elem().Field(0)
		_, err := excer.Exec(st.Context, v, st.req, st.opt)
		assert.Nil(t, err, testName)
		assert.Equal(t, st.expect, v.Interface(), testName)
	}
}

func TestHttpQuery(t *testing.T) {
	_, err := newHTTPExecer(bytes.Split([]byte("http.somehow"), dot))
	assert.NotNil(t, err)

	_, err = newHTTPExecer(bytes.Split([]byte("http.query.id.a"), dot))
	assert.NotNil(t, err)

	excer, err := newHTTPExecer(bytes.Split([]byte("http.query.id"), dot))
	assert.Nil(t, err)

	for testName, st := range map[string]struct {
		context.Context
		value  interface{}
		req    *http.Request
		opt    *DefaultOption
		expect interface{}
	}{
		"http-query-id": {
			context.Background(),
			struct{ T int }{},
			newReq().addQueryParam("id", "123").r(),
			nil,
			int(123),
		},
		"http-query-id-default": {
			context.Background(),
			struct{ T int }{},
			newReq().r(),
			&DefaultOption{IsDefaultExists: true, DefaultValue: "123456", DefaultSplitFlag: "|"},
			int(123456),
		},
	} {
		v := reflect.New(reflect.TypeOf(st.value)).Elem().Field(0)
		_, err := excer.Exec(st.Context, v, st.req, st.opt)
		assert.Nil(t, err, testName)
		assert.Equal(t, st.expect, v.Interface(), testName)
	}
}

func TestBaseType(t *testing.T) {
	err := errors.New("some error")
	for testName, tt := range map[string]struct {
		value  interface{}
		opt    *DefaultOption
		from   []string
		expect interface{}
	}{
		"base-type-int": {
			struct{ F int }{}, nil, []string{"9"}, int(9),
		},
		"base-type-int-float": {
			struct{ F int }{}, nil, []string{"9.9"}, err,
		},
		"base-type-int-string": {
			struct{ F int }{}, nil, []string{"abc"}, err,
		},

		"base-type-int8": {
			struct{ F int8 }{}, nil, []string{"9"}, int8(9),
		},
		"base-type-int8-float": {
			struct{ F int8 }{}, nil, []string{"9.9"}, err,
		},
		"base-type-int8-string": {
			struct{ F int8 }{}, nil, []string{"abc"}, err,
		},

		"base-type-int16": {
			struct{ F int16 }{}, nil, []string{"9"}, int16(9),
		},

		"base-type-int16-float": {
			struct{ F int16 }{}, nil, []string{"9.9"}, err,
		},
		"base-type-int16-string": {
			struct{ F int16 }{}, nil, []string{"abc"}, err,
		},

		"base-type-int32": {
			struct{ F int32 }{}, nil, []string{"9"}, int32(9),
		},
		"base-type-int32-float": {
			struct{ F int32 }{}, nil, []string{"9.9"}, err,
		},
		"base-type-int32-string": {
			struct{ F int32 }{}, nil, []string{"abc"}, err,
		},

		"base-type-int64": {
			struct{ F int64 }{}, nil, []string{"9"}, int64(9),
		},
		"base-type-int64-float": {
			struct{ F int64 }{}, nil, []string{"9.9"}, err,
		},
		"base-type-int64-string": {
			struct{ F int64 }{}, nil, []string{"abc"}, err,
		},

		"base-type-uint": {
			struct{ F uint }{}, nil, []string{"9"}, uint(9),
		},
		"base-type-uint-int": {
			struct{ F uint }{}, nil, []string{"-123"}, err,
		},
		"base-type-uint-float": {
			struct{ F uint }{}, nil, []string{"9.9"}, err,
		},
		"base-type-uint-string": {
			struct{ F uint }{}, nil, []string{"abc"}, err,
		},

		"base-type-uint8": {
			struct{ F uint8 }{}, nil, []string{"9"}, uint8(9),
		},
		"base-type-uint8-int": {
			struct{ F uint8 }{}, nil, []string{"-123"}, err,
		},
		"base-type-uint8-float": {
			struct{ F uint8 }{}, nil, []string{"9.9"}, err,
		},
		"base-type-uint8-string": {
			struct{ F uint8 }{}, nil, []string{"abc"}, err,
		},

		"base-type-uint16": {
			struct{ F uint16 }{}, nil, []string{"9"}, uint16(9),
		},
		"base-type-uint16-int": {
			struct{ F uint16 }{}, nil, []string{"-123"}, err,
		},
		"base-type-uint16-float": {
			struct{ F uint16 }{}, nil, []string{"9.9"}, err,
		},
		"base-type-uint16-string": {
			struct{ F uint16 }{}, nil, []string{"abc"}, err,
		},

		"base-type-uint32": {
			struct{ F uint32 }{}, nil, []string{"9"}, uint32(9),
		},
		"base-type-uint32-int": {
			struct{ F uint32 }{}, nil, []string{"-123"}, err,
		},
		"base-type-uint32-float": {
			struct{ F uint32 }{}, nil, []string{"9.9"}, err,
		},
		"base-type-uint32-string": {
			struct{ F uint32 }{}, nil, []string{"abc"}, err,
		},

		"base-type-uint64": {
			struct{ F uint64 }{}, nil, []string{"9"}, uint64(9),
		},
		"base-type-uint64-int": {
			struct{ F uint64 }{}, nil, []string{"-123"}, err,
		},
		"base-type-uint64-float": {
			struct{ F uint64 }{}, nil, []string{"9.9"}, err,
		},
		"base-type-uint64-string": {
			struct{ F uint64 }{}, nil, []string{"abc"}, err,
		},

		"base-type-bool": {
			struct{ F bool }{}, nil, []string{"true"}, true,
		},
		"base-type-bool-string": {
			struct{ F bool }{}, nil, []string{"abc"}, err,
		},
		"base-type-bool-int": {
			struct{ F bool }{}, nil, []string{"123"}, err,
		},

		"base-type-float32": {
			struct{ F float32 }{}, nil, []string{"9.9"}, float32(9.9),
		},
		"base-type-float32-string": {
			struct{ F float32 }{}, nil, []string{"abc"}, err,
		},

		"base-type-float64": {
			struct{ F float64 }{}, nil, []string{"9.9"}, float64(9.9),
		},
		"base-type-float64-string": {
			struct{ F float64 }{}, nil, []string{"abc"}, err,
		},

		"base-type-string": {
			struct{ F string }{}, nil, []string{"9"}, "9",
		},

		"base-type-duration": {
			struct{ F time.Duration }{}, nil, []string{"1m10s"}, time.Second * 70,
		},
		"base-type-duration-err": {
			struct{ F time.Duration }{}, nil, []string{"abc"}, err,
		},

		"base-type-slice-string": {
			struct{ F []string }{}, nil, []string{"9", "10", "11"}, []string{"9", "10", "11"},
		},

		"base-type-array-string": {
			struct{ F [3]string }{}, nil, []string{"9", "10", "11"}, [3]string{"9", "10", "11"},
		},

		"base-type-int-default": {
			struct{ F int }{}, &DefaultOption{
				IsDefaultExists:  true,
				DefaultValue:     "999",
				DefaultSplitFlag: "|",
			}, nil, int(999),
		},

		"base-type-uint-default": {
			struct{ F uint }{}, &DefaultOption{
				IsDefaultExists:  true,
				DefaultValue:     "9",
				DefaultSplitFlag: "|",
			}, nil, uint(9),
		},

		"base-type-string-default": {
			struct{ F string }{}, &DefaultOption{
				IsDefaultExists:  true,
				DefaultValue:     "9",
				DefaultSplitFlag: "|",
			}, nil, "9",
		},

		"base-type-bool-default": {
			struct{ F bool }{}, &DefaultOption{
				IsDefaultExists:  true,
				DefaultValue:     "true",
				DefaultSplitFlag: "|",
			}, nil, true,
		},

		"base-type-array-int-default": {
			struct{ F [3]int }{}, &DefaultOption{
				IsDefaultExists:  true,
				DefaultValue:     "1|2|3",
				DefaultSplitFlag: "|",
			}, nil, [3]int{1, 2, 3},
		},

		"base-type-slice-int-default": {
			struct{ F []int }{}, &DefaultOption{
				IsDefaultExists:  true,
				DefaultValue:     "1|2|3",
				DefaultSplitFlag: "|",
			}, nil, []int{1, 2, 3},
		},
	} {
		val := reflect.New(reflect.TypeOf(tt.value))
		val.Elem().Set(reflect.ValueOf(tt.value))
		f := val.Elem().Field(0)
		err := TrySet(f, tt.from, tt.opt)
		if _, ok := tt.expect.(error); ok {
			assert.NotNilf(t, err, testName)
		} else {
			assert.Nilf(t, err, testName)
			assert.Equal(t, tt.expect, f.Interface(), testName)
		}

	}
}
