# gbind包

## 版本说明

## 简介
	针对http请求封装通用的参数解析、参数校验逻辑，尽量减少日常开发的重复代码，`几行代码进行让解决所有参数的绑定和校验`

## 功能列表
- 针对http协议公共协议参数的绑定
	-	针对http uri参数进行绑定	
		- `gbind:"http.path"`
	-	针对http query参数进行绑定	
		- `gbind:"http.query.变量名"`
	-	针对http header参数进行绑定 
		- `gbind:"http.header.变量名"`
	-	针对http post参数进行绑定 	
		- `gbind:"http.post.变量名"`
	-	针对http cookie参数进行绑定 
		- `gbind:"http.cookie.变量名"`
-	针对body为json格式的统一遵循golang json解析格式
-	支持default标签设置默认值
	-	`gbind:"http.query.变量名,default=123"`
-	自定义绑定逻辑
	-	`gbind:"业务自定义标签"`
	-	业务可以根据需求进行标签自定义
	- 	自定义 `simple.key`的解析规则, demo

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

-	[参数校验逻辑参考validate包](https://pkg.go.dev/gopkg.in/go-playground/validator.v9	)
	- `gbind:"http.query.变量名,default=123" validate="required,lt=100"`
	-  支持自定义的解析规则, demo

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

-	通过定义 err_msg 的tag，在参数校验失败时支持自定义错误信息
	- `gbind:"http.cookie.BDUSS" validate="required,lt=100" err_msg="请想完成登录"`	

## 绑定支持的基础数据类型
- int
- int8
- int16
- int32
- int64
- uint
- uint8
- uint16
- uint32
- uint64
- float32
- float64
- string
- slice
- array

## 关于性能
- http的query+form参数绑定，gin、gbind包对比 <font color=#008000 >性能有大幅度提升</font>
```	
BenchmarkBind/gin-query-form-8         	  612357	      1937 ns/op	     304 B/op	      20 allocs/op
BenchmarkBind/gbind-query-form-8       	 6981271	      171.3 ns/op	     200 B/op	       5 allocs/op
```

- http的query+form+cookie参数绑定，gin、gbind包对比，<font color=#008000 >性能有大幅度提升</font>
```
BenchmarkBind/gin-query-form-header-8  	  232152	      5143 ns/op	     736 B/op	      53 allocs/op
BenchmarkBind/gbind-query-form-header-8   6673236	      180.0 ns/op	     232 B/op	       5 allocs/op	
```
## 使用demo

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
	Bduss  string `gbind:"http.cookie.BDUSS" validate:"required" err_msg:"please login"`
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
		Name:  "BDUSS",
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
	//	"Bduss": "foo-bar-andsoon",
	//	"Host": "gbind.baidu.com",
	//	"Uids": [
	//		1,
	//		2,
	//		3
	//	]
	//}
}
```