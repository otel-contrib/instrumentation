/*
Package http provides HTTP client and server implementations.
*/
package http

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// A Client is an HTTP client. Its zero value (DefaultClient) is a usable client that uses DefaultTransport.
type Client = http.Client

// DefaultClient is the default Client and is used by Get, Head, and Post.
var DefaultClient = NewClient(http.DefaultClient)

// RoundTripper is an interface representing the ability to execute a single HTTP transaction,
// obtaining the Response for a given Request.
type RoundTripper = http.RoundTripper

// NewClient returns a client that provides OpenTelemetry tracing and metrics.
func NewClient(c *Client, opts ...Option) *Client {
	transport := c.Transport
	if transport == nil {
		transport = DefaultTransport
	}

	transport, err := NewOTelTransport(transport, opts...)
	if err != nil {
		panic(err)
	}
	c.Transport = transport

	return c
}

// Get issues a GET to the specified URL.
func Get(url string) (resp *Response, err error) {
	return DefaultClient.Get(url)
}

// GetWithContext is a convenient replacement for http.Get that adds a span around the request.
func GetWithContext(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := NewRequestWithContext(ctx, MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return DefaultClient.Do(req)
}

// Post issues a POST to the specified URL.
func Post(url, contentType string, body io.Reader) (resp *Response, err error) {
	return DefaultClient.Post(url, contentType, body)
}

// PostWithContext is a convenient replacement for http.Post that adds a span around the request.
func PostWithContext(ctx context.Context, url, contentType string, body io.Reader) (resp *Response, err error) {
	req, err := NewRequest(MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return DefaultClient.Do(req)
}

// PostForm issues a POST to the specified URL, with data's keys and values URL-encoded as the request body.
func PostForm(url string, data url.Values) (resp *Response, err error) {
	return DefaultClient.PostForm(url, data)
}

// PostFormWithContext is a convenient replacement for http.PostForm that adds a span around the request.
func PostFormWithContext(ctx context.Context, url string, data url.Values) (resp *Response, err error) {
	return PostWithContext(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

// Head issues a HEAD to the specified URL.
func Head(url string) (resp *Response, err error) {
	return DefaultClient.Head(url)
}

// HeadWithContext is a convenient replacement for http.Head that adds a span around the request.
func HeadWithContext(ctx context.Context, url string) (resp *Response, err error) {
	req, err := NewRequest(MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	return DefaultClient.Do(req)
}
