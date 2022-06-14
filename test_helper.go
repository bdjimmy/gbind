package gbind

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// hr helper to create http.Request
type hr struct {
	method      string
	path        string
	header      http.Header
	queryParams url.Values
	formParams  url.Values
	cookies     []*http.Cookie
	body        io.ReadCloser
}

// newReq create a pointer to http.Request
func newReq() *hr {
	return &hr{
		path:        "/api/test",
		queryParams: url.Values{},
		formParams:  url.Values{},
		header:      make(http.Header),
		cookies:     make([]*http.Cookie, 0),
	}
}

// setMethod set the request method
func (hr *hr) setMethod(method string) *hr {
	hr.method = method
	return hr
}

// setPath set the request path
func (hr *hr) setPath(path string) *hr {
	hr.path = path
	return hr
}

// addQueryParam add the request query param
func (hr *hr) addQueryParam(k, v string) *hr {
	hr.queryParams.Add(k, v)
	return hr
}

// addHeader add the request header
func (hr *hr) addHeader(k, v string) *hr {
	hr.header.Add(k, v)
	return hr
}

// addCookie add the request cookie
func (hr *hr) addCookie(k, v string) *hr {
	hr.cookies = append(hr.cookies, &http.Cookie{
		Name:  k,
		Value: v,
	})
	return hr
}

// addFormParam add the request form param
func (hr *hr) addFormParam(k, v string) *hr {
	hr.formParams.Add(k, v)
	return hr
}

// setBody set the request body
func (hr *hr) setBody(body string) *hr {
	hr.body = io.NopCloser(strings.NewReader(body))
	return hr
}

// R translate into http.Request
func (hr *hr) r() *http.Request {
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
