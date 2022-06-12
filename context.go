package gbind

import (
	"context"
	"net/http"
	"net/url"
)

var (
	defaultMultipartMemory = int64(32 << 20) // 32 MB
)

type metaKey struct{}

// HTTPMetaData http execer
type HTTPMetaData struct {
	request    *http.Request
	queryCache url.Values
	formCache  url.Values
}

func newHTTPContext(ctx context.Context, req *http.Request) context.Context {
	_, ok := ctx.Value(metaKey{}).(*HTTPMetaData)
	if !ok {
		md := &HTTPMetaData{
			request: req,
		}
		ctx = context.WithValue(ctx, metaKey{}, md)
	}
	return ctx
}

func mustContextHTTPMeta(ctx context.Context) *HTTPMetaData {
	md, ok := ctx.Value(metaKey{}).(*HTTPMetaData)
	if !ok {
		panic("context should use gbind.NewContext to initial first")
	}
	return md
}

func (hm *HTTPMetaData) initQueryCache() {
	if hm.queryCache == nil {
		if hm.request == nil {
			hm.queryCache = url.Values{}
		} else {
			hm.queryCache = hm.request.URL.Query()
		}
	}
}

func (hm *HTTPMetaData) initFormCache() {
	if hm.formCache == nil {
		hm.request.ParseMultipartForm(defaultMultipartMemory)
		hm.formCache = hm.request.PostForm
	}
}

func (hm *HTTPMetaData) getFormArray(key string) (values []string) {
	hm.initFormCache()
	return hm.formCache[key]
}

func (hm *HTTPMetaData) getQueryArray(key string) (values []string) {
	hm.initQueryCache()
	return hm.queryCache[key]
}
