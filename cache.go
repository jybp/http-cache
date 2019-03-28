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
}

// Default uses http.DefaultTransport to make actual requests.
func Default(c Cache) *Transport {
	return Custom(http.DefaultTransport, c)
}

// Custom uses t to make actual requests.
func Custom(t http.RoundTripper, c Cache) *Transport {
	return &Transport{Transport: t, Cache: c}
}

// RoundTrip implements the cache logic.
func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
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
