# gbind包
[![Go Report Card](https://goreportcard.com/badge/github.com/bdjimmy/gbind)](https://goreportcard.com/report/github.com/bdjimmy/gbind)
<a title="Doc for ants" target="_blank" href="https://pkg.go.dev/github.com/bdjimmy/gbind?tab=doc"><img src="https://img.shields.io/badge/go.dev-doc-007d9c?style=flat-square&logo=read-the-docs" /></a>
<a title="Codecov" target="_blank" href="https://codecov.io/gh/bdjimmy/gbind"><img src="https://img.shields.io/codecov/c/github/bdjimmy/gbind?style=flat-square&logo=codecov" /></a>

	封装通用的参数解析、参数校验逻辑，尽量减少日常开发的重复代码，几行代码进行解决所有参数的绑定和校验


## Features
+ 根据tag信息将数据绑定到指定的结构体
	- 内置HTTP request的path、query、form、header、cookie的绑定能力
		- 针对http uri参数进行绑定,  `gbind:"http.path"`
		- 针对http query参数进行绑定 `gbind:"http.query.变量名"`
		- 针对http header参数进行绑定  `gbind:"http.header.变量名"`
		- 针对http form参数进行绑定 `gbind:"http.form.变量名"`
		- 针对http cookie参数进行绑定 `gbind:"http.cookie.变量名"`
	- 内置json的绑定能力，借助encoding/json实现
		- 针对body为json格式的统一遵循golang json解析格式 `json:"name"`
	- 支持设置绑定字段的默认值
		- 在没有传入数据时，支持设置绑定字段的默认值 `gbind:"http.query.变量名,default=123"`
	- 支持自定义绑定解析逻辑（不仅仅局限于针对HTTP request，使用gbind可以做类似数据库tag等场景的绑定）
		- 通过调用 `RegisterBindFunc` 函数可以注册自定义的绑定逻辑，例如实现 `gbind:"simple.key"` 形式的绑定

- 根据tag信息进行字段值的校验，[参数校验逻辑参考validate包](https://pkg.go.dev/gopkg.in/go-playground/validator.v9	)
	- 根据定义的 `validate`tag进行绑定字段的数据校验，依赖github.com/go-playground/validator实现， `validate="required,lt=100"`
	- 支持自定义校验逻辑，通过调用 `RegisterCustomValidation`函数可以自定义数据校验逻辑
	- 支持自定义校验失败的错误提示信息
		- 通过定义 err_msg 的tag，在参数校验失败时支持自定义错误信息，demo `gbind:"http.cookie.Token" validate="required,lt=100" err_msg="请想完成登录"`
## Usage example
- 使用gbind的web API请求参数进行绑定和校验

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
- 自定义绑定逻辑，可以实现不同场景的绑定，demo `gbind:"simple.key"`
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
- 自定义数据校验逻辑，可以实现不同场景的校验，demo `validate:"is-awesome"`
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
+ 针对gin框架、gbind包的绑定能力进行了压测，query+form简单绑定能力，gbind有**10倍以上**的性能提升，query+form+header复杂的绑定能力，gbind有**30倍以上**的性能提升，具体数据如下
	- http的query+form参数绑定，gin、gbind包对比
```	
BenchmarkBind/gin-query-form-8         	  612357	      1937 ns/op	     304 B/op	      20 allocs/op
BenchmarkBind/gbind-query-form-8       	 6981271	      171.3 ns/op	     200 B/op	       5 allocs/op
```

	- http的query+form+cookie参数绑定，gin、gbind包对比
```
BenchmarkBind/gin-query-form-header-8  	  232152	      5143 ns/op	     736 B/op	      53 allocs/op
BenchmarkBind/gbind-query-form-header-8   6673236	      180.0 ns/op	     232 B/op	       5 allocs/op	
```

## 绑定支持的基础数据类型
+ 基础数据类型
	- int、int8、int16、int32、int64
	- uint、uint8、uint16、uint32、uint64
	- float32、float64
	- bool
	- string
+ 其他数据类型	
	- ptr
		- *int、*uint、*float32、*string
		- **int、**uint、**float32、**string
		- 任何基础数据类型的多级指针
	- slice
		- []int、[]uint、[]bool、[]string等
		- []*int、[]*uint、[]*bool、[]*string等
		- 任何基础数据类型（包含指针）的切片
	- array
		- [1]int、[2]uint、[3]bool、[4]string等
		- [5]*int、[6]*uint、[7]*bool、[8]*string等
		- 任何基础数据类型（包含指针）的数组
	- time.Duration	

