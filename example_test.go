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
	Bduss  string `gbind:"http.cookie.BDUSS" validate:"required" err_msg:"please complete the login operation first"`
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
	// {
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
	// }
}
