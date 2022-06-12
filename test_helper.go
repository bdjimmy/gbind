package gbind

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HR helper to create http.Request
type HR struct {
	method      string
	path        string
	header      http.Header
	queryParams url.Values
	formParams  url.Values
	cookies     []*http.Cookie
	body        io.ReadCloser
}

// NewReq create a pointer to http.Request
func NewReq() *HR {
	return &HR{
		path:        "/api/test",
		queryParams: url.Values{},
		formParams:  url.Values{},
		header:      make(http.Header),
		cookies:     make([]*http.Cookie, 0),
	}
}

// SetMethod set the request method
func (hr *HR) SetMethod(method string) *HR {
	hr.method = method
	return hr
}

// SetPath set the request path
func (hr *HR) SetPath(path string) *HR {
	hr.path = path
	return hr
}

// AddQueryParam add the request query param
func (hr *HR) AddQueryParam(k, v string) *HR {
	hr.queryParams.Add(k, v)
	return hr
}

// AddHeader add the request header
func (hr *HR) AddHeader(k, v string) *HR {
	hr.header.Add(k, v)
	return hr
}

// AddCookie add the request cookie
func (hr *HR) AddCookie(k, v string) *HR {
	hr.cookies = append(hr.cookies, &http.Cookie{
		Name:  k,
		Value: v,
	})
	return hr
}

// AddFormParam add the request form param
func (hr *HR) AddFormParam(k, v string) *HR {
	hr.formParams.Add(k, v)
	return hr
}

// SetBody set the request body
func (hr *HR) SetBody(body string) *HR {
	hr.body = io.NopCloser(strings.NewReader(body))
	return hr
}

// R translate into http.Request
func (hr *HR) R() *http.Request {
	u, e := url.Parse(fmt.Sprintf("http://www.test.com%s?%s", hr.path, hr.queryParams.Encode()))
	if e != nil {
		panic(e)
	}
	req := &http.Request{
		Method:   hr.method,
		Header:   hr.header,
		PostForm: hr.formParams,
		URL:      u,
		Body:     hr.body,
	}
	for _, c := range hr.cookies {
		req.AddCookie(c)
	}
	return req
}
