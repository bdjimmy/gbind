package gbind

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type BindTestQuery struct {
	Ak string `form:"ak,default=ak-default" gbind:"http.query.ak,default=ak-default"`
	Tk string `form:"tk" gbind:"http.query.tk"`
	Ts int64  `form:"ts" gbind:"http.query.ts"`
}

type BindTestForm struct {
	Page   int    `form:"page" gbind:"http.post.page"`
	Size   int    `form:"size" gbind:"http.post.size"`
	Appkey string `form:"appkey" gbind:"http.post.appkey"`
}

type BindTestHeader struct {
	Host string `header:"host" gbind:"http.header.host"`
}

type BindTestCookie struct {
	Bduss string `gbind:"http.cookie.BDUSS"`
}

func BenchmarkBind(b *testing.B) {
	req := NewReq().
		AddQueryParam("ak", "ak1").
		AddQueryParam("tk", "tk1").
		AddQueryParam("ts", "123456789"). //
		AddFormParam("page", "1").
		AddFormParam("size", "2").
		AddFormParam("appkey", "3").
		AddHeader("host", "www.baidu.com").R()
	req.AddCookie(&http.Cookie{
		Name:  "BDUSS",
		Value: "bduss-value",
	})

	{
		type BindTest struct {
			BindTestQuery
			BindTestForm
		}

		b.ResetTimer()
		b.Run("gin-query-form", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				ginCtx := gin.Context{
					Request: req,
				}
				value := &BindTest{}
				ginCtx.ShouldBind(&value)
			}
		})

		b.ResetTimer()
		b.Run("gbind-query-form", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				value := &BindTest{}
				Bind(context.Background(), &value, req)
			}
		})
	}

	{
		type BindTest struct {
			BindTestQuery
			BindTestForm
			BindTestHeader
			BindTestCookie
		}

		b.ResetTimer()
		b.Run("gin-query-form-header", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				ginCtx := gin.Context{
					Request: req,
				}
				value := &BindTest{}
				ginCtx.ShouldBind(&value)
				ginCtx.ShouldBindHeader(&value)
			}
		})

		b.ResetTimer()
		b.Run("gbind-query-form-header", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				value := &BindTest{}
				Bind(context.Background(), &value, req)
			}
		})
	}
}
