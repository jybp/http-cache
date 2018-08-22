// Package httpcache provides a simple http.RoundTripper to cache http responses.
package httpcache

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/pkg/errors"
)

// Cache describes how to retrieve and store responses.
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, response []byte)
}

// Transport implements http.RoundTripper and returns responses from a cache.
type Transport struct {
	Transport http.RoundTripper // used to make actual requests.
	Cache     Cache
	Filter    func(r *http.Request) bool // return true to ignore the cache.
}

// RoundTrip implements the cache logic.
// If the request is filtered out, the cache is not used and the request is executed.
// If the cache contains a cached response, it will be used without executing the request.
func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.Filter != nil && t.Filter(r) {
		return t.Transport.RoundTrip(r)
	}
	h := md5.New()
	io.WriteString(h, r.URL.String())
	key := hex.EncodeToString(h.Sum(nil))
	b, ok := t.Cache.Get(key)
	if !ok {
		transport := t.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		resp, err := transport.RoundTrip(r)
		if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return resp, errors.WithStack(err)
		}
		b, err = httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		t.Cache.Set(key, b)
	}
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(b)), r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp, nil
}
