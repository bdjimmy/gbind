English | [🇨🇳中文](README_ZH.md)
# gbind
	Encapsulate general parameter parsing and parameter verification logic, 
	minimize repetitive code in daily development, and solve parameter binding 
	and verification in a few lines of code


## Features
+ Bind data to the specified structure based on tag information
	- Built-in HTTP request path, query, form, header, cookie binding ability
		- Binding for http uri parameters  `gbind:"http.path"`
		- Binding for http query parameters `gbind:"http.query.varname"`
		- Binding for http header parameters  `gbind:"http.header.varname"`
		- Binding for http form parameters `gbind:"http.form.varname"`
		- Binding for http cookie parameters `gbind:"http.cookie.varname"`
	- Built-in json binding capability, implemented with encoding/json
		- For HTTP body in json format, follow golang json parsing format uniformly `json:"varname"`
	- Support for setting default values of bound fields
		- Supports setting default values of bound fields when no data is passed in `gbind:"http.query.varname,default=123"`
	- Support custom binding parsing logic (not limited to HTTP requests, using gbind can do bindings similar to database tags and other scenarios)
		- You can register custom binding logic by calling the `RegisterBindFunc` function, such as implementing a binding of the form `gbind:"simple.key"`

- Validate the field value according to the tag information, [parameter validation logic refer to the validate package](https://pkg.go.dev/gopkg.in/go-playground/validator.v9	)
	- Data validation of bound fields is performed according to the defined `validate`tag, which depends on github.com/go-playground/validator implementation, `validate="required,lt=100"`
	- Support custom validation logic, you can customize the data validation logic by calling the `RegisterCustomValidation` function
	- Support custom error message for validation failure
	- By defining the tag of err_msg, it supports custom error message when parameter validation fails, demo `gbind:"http.cookie.Token" validate="required,lt=100" err_msg="Please complete the login"`
## Usage example
- Use gbind's web API request parameters for binding and verification

```golang
package gbind

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type Params struct {
	API    string `gbind:"http.path,default=/api/test"`
	Appkey string `gbind:"http.query.appkey,default=appkey-default"`
	Page   int    `gbind:"http.query.page,default=1"`
	Size   int    `gbind:"http.query.size,default=10"`
	Token  string `gbind:"http.cookie.Token" validate:"required" err_msg:"please login"`
	Host   string `gbind:"http.header.host,default=www.baidu.com"`
	Uids   []int  `gbind:"http.form.uids"`
}

func Controller(w http.ResponseWriter, r *http.Request) {
	var requestParams = &Params{}
	if _, err := BindWithValidate(context.Background(), requestParams, r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bs, _ := json.MarshalIndent(requestParams, "", "\t")
	w.Write(bs)
}

func ExampleGbind() {
	w := httptest.NewRecorder()
	u, _ := url.Parse("http://gbind.baidu.com/api/test?appkey=abc&page=2")
	r := &http.Request{
		Method: http.MethodPost,
		Header: map[string][]string{
			"Host": {"gbind.baidu.com"},
		},
		PostForm: url.Values{
			"uids": {"1", "2", "3"},
		},
		URL: u,
	}
	r.AddCookie(&http.Cookie{
		Name:  "Token",
		Value: "foo-bar-andsoon",
	})

	Controller(w, r)

	fmt.Println(w.Result().Status)
	fmt.Println(w.Body.String())

	// Output:
	// 200 OK
	//{
	//	"API": "/api/test",
	//	"Appkey": "abc",
	//	"Page": 2,
	//	"Size": 10,
	//	"Token": "foo-bar-andsoon",
	//	"Host": "gbind.baidu.com",
	//	"Uids": [
	//		1,
	//		2,
	//		3
	//	]
	//}
}
```
- Customize the binding logic, you can realize the binding of different scenarios, demo `gbind:"simple.key"`
```golang
package gbind

func TestRegisterBindFunc(t *testing.T) {
	g := NewGbind()
	g.RegisterBindFunc("simple", NewSimpleExecer)

	type Foo struct {
		Key string `gbind:"simple.key"`
	}
	f := &Foo{}
	_, err := g.Bind(context.WithValue(context.Background(), exprKey{}, "simple-k-d"), f, nil)
	assert.Nil(t, err)
	assert.Equal(t, "simple-k-d", f.Key)	
}

type exprKey struct{}

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

```
- Custom data verification logic, can realize verification in different scenarios, demo `validate:"is-awesome"`
```golang
func TestRegisterCustomValidation(t *testing.T) {
	g := NewGbind()

	g.RegisterCustomValidation("is-awesome", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "awesome"
	})

	{
		type Foo struct {
			Appkey string `gbind:"http.query.appkey" validate:"is-awesome"`
		}
		f := &Foo{}
		req := NewReq().AddQueryParam("appkey", "awesome").R()
		_, err := g.BindWithValidate(context.Background(), f, req)
		assert.Nil(t, err)
	}
}
```

## benchmark
+ Stressed the binding capabilities of the gin framework and the gbind package. The simple binding capabilities of query+form, gbind has a performance improvement of more than **10 times**, and the complex binding capabilities of query+form+header, gbind has **30 times or more** performance improvement, the specific data are as follows
	- HTTP query+form parameter binding, gin, gbind package comparison
```	
BenchmarkBind/gin-query-form-8         	  612357	      1937 ns/op	     304 B/op	      20 allocs/op
BenchmarkBind/gbind-query-form-8       	 6981271	      171.3 ns/op	     200 B/op	       5 allocs/op
```

	- HTTP query+form+cookie parameter binding, gin, gbind package comparison
```
BenchmarkBind/gin-query-form-header-8  	  232152	      5143 ns/op	     736 B/op	      53 allocs/op
BenchmarkBind/gbind-query-form-header-8   6673236	      180.0 ns/op	     232 B/op	       5 allocs/op	
```

## Binding supported underlying data types
+ basic data type
	- int、int8、int16、int32、int64
	- uint、uint8、uint16、uint32、uint64
	- float32、float64
	- bool
	- string
+ other data types	
	- ptr
		- *int、*uint、*float32、*string
		- **int、**uint、**float32、**string
		- Multilevel pointer to any underlying data type
	- slice
		- []int, []uint, []bool, []string, etc.
		- []*int, []*uint, []*bool, []*string, etc.
		- a slice of any underlying data type (including pointers)
	- array
		- [1]int, [2]uint, [3]bool, [4]string, etc.
		- [5]*int, [6]*uint, [7]*bool, [8]*string, etc.
		- Arrays of any underlying data type (including pointers)
		- time.Duration

