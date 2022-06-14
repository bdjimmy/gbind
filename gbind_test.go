package gbind

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestBasicStruct(t *testing.T) {
	type Foo struct {
		A int
		B int
		C int
	}
	f := &Foo{1, 2, 3}
	fv := reflect.ValueOf(f)
	st, err := defaultGbind.compile(f)
	assert.Nil(t, err)

	v := st.getFieldByNS(fv, "Foo.A")
	assert.Equal(t, f.A, v.Interface().(int))

	v = st.getFieldByNS(fv, "Foo.B")
	assert.Equal(t, f.B, v.Interface().(int))

	v = st.getFieldByNS(fv, "Foo.C")
	assert.Equal(t, f.C, v.Interface().(int))
}

func TestEmbeddedStruct(t *testing.T) {
	type Foo struct {
		A int
	}

	type Bar struct {
		Foo
		B int
		C int
	}

	type Baz struct {
		A int
		Bar
	}
	z := &Baz{
		A: 1,
		Bar: Bar{
			Foo: Foo{
				A: 2,
			},
			B: 3,
			C: 40,
		},
	}

	st, err := defaultGbind.compile(z)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(st.fields))

	zv := reflect.ValueOf(z)

	v := st.getFieldByNS(zv, "Baz.A")
	assert.Equal(t, z.A, v.Interface().(int))

	v = st.getFieldByNS(zv, "Baz.Bar.B")
	assert.Equal(t, z.Bar.B, v.Interface().(int))

	v = st.getFieldByNS(zv, "Baz.Bar.C")
	assert.Equal(t, z.Bar.C, v.Interface().(int))

	v = st.getFieldByNS(zv, "Baz.Bar.Foo.A")
	assert.Equal(t, z.Bar.Foo.A, v.Interface().(int))
}

func TestInlineStruct(t *testing.T) {
	type Foo struct {
		Name string
		ID   int
	}
	type Bar Foo
	type baz struct {
		Foo
		Bar
	}
	b := &baz{
		Foo: Foo{
			Name: "Joe", ID: 2,
		},
		Bar: Bar{
			Name: "Dick", ID: 1,
		},
	}
	bv := reflect.ValueOf(b)

	st, err := defaultGbind.compile(b)
	assert.Nil(t, err)

	v := st.getFieldByNS(bv, "baz.Bar.ID")
	assert.Equal(t, b.Bar.ID, v.Interface().(int))

	v = st.getFieldByNS(bv, "baz.Bar.Name")
	assert.Equal(t, b.Bar.Name, v.Interface().(string))

	v = st.getFieldByNS(bv, "baz.Foo.ID")
	assert.Equal(t, b.Foo.ID, v.Interface().(int))

	v = st.getFieldByNS(bv, "baz.Foo.Name")
	assert.Equal(t, b.Foo.Name, v.Interface().(string))

}

func TestFieldByIndexs(t *testing.T) {
	type Foo struct {
		A int
	}

	type Bar struct {
		Foo
		B int
		C *int
	}

	type Baz struct {
		A int
		Bar
	}
	four := 4
	b := &Baz{
		A: 1,
		Bar: Bar{
			Foo: Foo{
				A: 2,
			},
			B: 3,
			C: &four,
		},
	}

	bv := reflect.ValueOf(b)

	v := fieldByIndexs(bv, []int{0})
	assert.Equal(t, 1, v.Interface().(int))

	v = fieldByIndexs(bv, []int{1, 0, 0})
	assert.Equal(t, 2, v.Interface().(int))

	v = fieldByIndexs(bv, []int{1, 1})
	assert.Equal(t, 3, v.Interface().(int))

	v = fieldByIndexs(bv, []int{1, 2})
	assert.Equal(t, 4, v.Interface().(int))

}

func TestBasicBind(t *testing.T) {
	// bind
	{
		type Foo struct {
			Appkey string `gbind:"http.query.appkey"`
		}
		f := &Foo{}
		req := newReq().addQueryParam("appkey", "abc").r()
		_, err := Bind(context.Background(), f, req)
		assert.Nil(t, err)
		assert.Equal(t, "abc", f.Appkey)
	}

	// bind with pointer
	{
		type Foo struct {
			Appkey *string `gbind:"http.query.appkey"`
		}
		f := &Foo{}
		req := newReq().addQueryParam("appkey", "abc").r()
		_, err := Bind(context.Background(), f, req)
		assert.Nil(t, err)
		assert.Equal(t, "abc", *(f.Appkey))
	}

	// default
	{
		type Foo struct {
			Appkey string `gbind:"http.query.appkey,default=123"`
		}
		f := &Foo{}
		req := newReq().r()
		_, err := Bind(context.Background(), f, req)
		assert.Nil(t, err)
		assert.Equal(t, "123", f.Appkey)
	}

	// validate
	{
		type Foo struct {
			Appkey string `gbind:"http.query.appkey" validate:"required"`
		}
		f := &Foo{}
		req := newReq().r()
		_, err := BindWithValidate(context.Background(), f, req)
		assert.NotNil(t, err)
	}

	// err_msg
	{
		type Foo struct {
			Appkey string `gbind:"http.query.appkey" validate:"required" err_msg:"field appkey is required"`
		}
		f := &Foo{}
		req := newReq().r()
		_, err := BindWithValidate(context.Background(), f, req)
		assert.NotNil(t, err)
		assert.Equal(t, "field appkey is required", err.Error())
	}
}

func TestBindTag(t *testing.T) {
	g := NewGbind(WithBindTag("mybind"))

	{
		type Foo struct {
			Appkey string `mybind:"http.query.appkey"`
		}
		f := &Foo{}
		req := newReq().addQueryParam("appkey", "abc").r()
		_, err := g.Bind(context.Background(), f, req)
		assert.Nil(t, err)
		assert.Equal(t, "abc", f.Appkey)
	}
}

func TestErrTag(t *testing.T) {
	g := NewGbind(WithErrTag("error"))

	// err_msg
	{
		type Foo struct {
			Appkey string `gbind:"http.query.appkey" validate:"required" error:"field appkey is required"`
		}
		f := &Foo{}
		req := newReq().r()
		_, err := g.BindWithValidate(context.Background(), f, req)
		assert.NotNil(t, err)
		assert.Equal(t, "field appkey is required", err.Error())
	}
}

func TestDefaultSplit(t *testing.T) {
	g := NewGbind(WithDefaultSplitFlag("-"))

	{
		type Foo struct {
			Uids []int `gbind:"http.query.uid,default=1-2-3"`
		}
		f := &Foo{}
		req := newReq().r()
		_, err := g.Bind(context.Background(), f, req)
		assert.Nil(t, err)
		assert.Equal(t, []int{1, 2, 3}, f.Uids)
	}
}

func TestRegisterBindFunc(t *testing.T) {
	g := NewGbind()
	g.RegisterBindFunc("simple", NewSimpleExecer)

	{
		type Foo struct {
			Key string `gbind:"simple.key"`
		}
		f := &Foo{}
		_, err := g.Bind(context.WithValue(context.Background(), exprKey{}, "simple-k-d"), f, nil)
		assert.Nil(t, err)
		assert.Equal(t, "simple-k-d", f.Key)
	}
}

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
		req := newReq().addQueryParam("appkey", "awesome").r()
		_, err := g.BindWithValidate(context.Background(), f, req)
		assert.Nil(t, err)
	}
}

func TestJosn(t *testing.T) {
	{
		type Foo struct {
			Appkey  string `json:"appkey"`
			AppName string `json:"appname"`
		}
		f := &Foo{}

		req := newReq().setBody(`{"appkey":"abc","appname":"123"}`).r()
		_, err := Bind(context.Background(), f, req)

		assert.Nil(t, err)
		assert.Equal(t, "abc", f.Appkey)
		assert.Equal(t, "123", f.AppName)
	}
	{
		type Foo struct {
			ID interface{} `json:"id"`
		}
		f := &Foo{}

		req := newReq().setBody(`{"id":280123412341234123}`).r()
		g := NewGbind(WithUseNumberForJSON(false))
		_, err := g.Bind(context.Background(), f, req)
		assert.Nil(t, err)
		assert.Equal(t, float64(280123412341234123), f.ID)

		req = newReq().setBody(`{"id":280123412341234123}`).r()
		g = NewGbind(WithUseNumberForJSON(true))
		_, err = g.Bind(context.Background(), f, req)
		assert.Nil(t, err)
		assert.Equal(t, json.Number("280123412341234123"), f.ID)
	}
}

func TestCheckValid(t *testing.T) {
	type Foo struct{}
	{
		var f = Foo{}
		err := NewGbind().checkValid(reflect.ValueOf(f))
		assert.NotNil(t, err)
	}
	{
		var f = Foo{}
		err := NewGbind().checkValid(reflect.ValueOf(&f))
		assert.Nil(t, err)
	}
	{
		var f = &Foo{}
		err := NewGbind().checkValid(reflect.ValueOf(&f))
		assert.NotNil(t, err)
	}
	{
		var f = 123
		err := NewGbind().checkValid(reflect.ValueOf(&f))
		assert.NotNil(t, err)
	}
}
