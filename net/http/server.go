package http

import "net/http"

// A Handler responds to an HTTP request.
type Handler = http.Handler

// The HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler that calls f.
type HandlerFunc = http.HandlerFunc
