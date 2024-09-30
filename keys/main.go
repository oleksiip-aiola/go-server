package keys

import "net/http"

// Define the contextKey type and the httpRequestKey
type ContextKey string

type HttpRequestResponse struct {
	Request  *http.Request
	Response http.ResponseWriter
}

// Shared key
const HttpRequestKey ContextKey = "httpRequest"
const HttpResponseWriterKey ContextKey = "httpResponseWriter"
