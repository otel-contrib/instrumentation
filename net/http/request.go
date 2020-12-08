/*
Package http provides HTTP client and server implementations.
*/
package http

import (
	"context"
	"io"
	"net/http"
)

// A Request represents an HTTP request received by a server or to be sent by a client.
type Request = http.Request

// NewRequest wraps NewRequestWithContext using the background context.
func NewRequest(method, url string, body io.Reader) (*Request, error) {
	return http.NewRequest(method, url, body)
}

// NewRequestWithContext returns a new Request given a method, URL, and optional body.
func NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*Request, error) {
	return http.NewRequestWithContext(ctx, method, url, body)
}
