// Package main contains the static base HTTP client code.
//
// This file holds the base HTTP client implementation that will be copied
// as-is to the generated client package. It's kept as a string constant
// to be embedded in the plugin binary.
package main

// httpClientBaseCode is the static base HTTP client implementation.
//
// This code is copied verbatim to the generated http/client.go file.
// It provides the core HTTP functionality that all service wrappers use.
//
// Key features:
// - JSON marshaling/unmarshaling with protojson
// - Bearer token authentication
// - GET and POST support
// - Consistent error handling
const httpClientBaseCode = `// Package http provides a reusable HTTP client for JSON-encoded proto messages.
//
// This client is designed for use with generated service wrappers.
// It handles:
// - JSON marshaling/unmarshaling with protojson
// - Bearer token authentication
// - Consistent error handling
// - Support for GET and POST operations
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Supported HTTP methods (only GET and POST are currently implemented)
const (
	MethodGET  = "GET"
	MethodPOST = "POST"
)

// HTTPClient wraps the standard http.Client with proto+JSON support.
//
// This is the core building block for service-specific clients.
// It's intentionally minimal and designed for code generation.
type HTTPClient struct {
	// BaseURL is the API base URL (e.g., "https://data.sandbox.iniciador.com.br")
	BaseURL string

	// HTTPClient is the underlying HTTP client (configure timeout, transport, etc.)
	HTTPClient *http.Client

	// Token is the Bearer token for authentication
	Token string
}

// Post sends a POST request with a JSON-encoded proto message body.
//
// The request is marshaled to JSON using protojson with camelCase field names.
// The response is unmarshaled from JSON back to a proto message.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - path: API path (e.g., "/v1/data/auth")
//   - req: Proto message to send as JSON body
//   - resp: Proto message to unmarshal response into
//
// Returns an error if the request fails or the response status is not 2xx.
func (c *HTTPClient) Post(ctx context.Context, path string, req proto.Message, resp proto.Message) error {
	return c.do(ctx, "POST", path, req, resp, "")
}

// PostWithWrap sends a POST request and wraps the response into a specified field before unmarshaling.
//
// This is useful when the API returns a plain array but proto requires a message wrapper.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - path: API path (e.g., "/v1/data/participants")
//   - req: Proto message to send as JSON body
//   - resp: Proto message to unmarshal response into
//   - wrapField: Field name to wrap the response array into (e.g., "response")
//
// Returns an error if the request fails or the response status is not 2xx.
func (c *HTTPClient) PostWithWrap(ctx context.Context, path string, req proto.Message, resp proto.Message, wrapField string) error {
	return c.do(ctx, "POST", path, req, resp, wrapField)
}

// Get sends a GET request and unmarshals the JSON response.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - path: API path (e.g., "/v1/data/links/123")
//   - resp: Proto message to unmarshal response into
//
// Returns an error if the request fails or the response status is not 2xx.
func (c *HTTPClient) Get(ctx context.Context, path string, resp proto.Message) error {
	return c.do(ctx, "GET", path, nil, resp, "")
}

// GetWithWrap sends a GET request and wraps the response into a specified field before unmarshaling.
//
// This is useful when the API returns a plain array but proto requires a message wrapper.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - path: API path (e.g., "/v1/data/participants")
//   - resp: Proto message to unmarshal response into
//   - wrapField: Field name to wrap the response array into (e.g., "response")
//
// Returns an error if the request fails or the response status is not 2xx.
func (c *HTTPClient) GetWithWrap(ctx context.Context, path string, resp proto.Message, wrapField string) error {
	return c.do(ctx, "GET", path, nil, resp, wrapField)
}

// do performs the actual HTTP request with proto message marshaling.
//
// This is the core method that handles:
// 1. Request marshaling (proto → JSON with camelCase)
// 2. HTTP request execution with proper headers
// 3. Response wrapping (if wrapField is specified)
// 4. Response unmarshaling (JSON → proto)
// 5. Error handling
//
// Parameters:
//   - wrapField: If non-empty, wraps the response JSON into this field name before unmarshaling
func (c *HTTPClient) do(ctx context.Context, method, path string, req proto.Message, resp proto.Message, wrapField string) error {
	// Validate HTTP method (only GET and POST are currently supported)
	if method != MethodGET && method != MethodPOST {
		return fmt.Errorf("unsupported HTTP method: %s (only GET and POST are supported)", method)
	}

	url := c.BaseURL + path

	// Marshal request body if provided
	var body io.Reader
	if req != nil {
		reqBytes, err := protojson.MarshalOptions{
			UseProtoNames:   false, // Use JSON names (camelCase: clientId, accessToken)
			EmitUnpopulated: false, // Don't include zero values
		}.Marshal(req)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(reqBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	if req != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.Token)
	}

	// Execute request
	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Check HTTP status
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		// Try to parse error response for better error messages
		var errResp map[string]interface{}
		json.Unmarshal(respBytes, &errResp)
		return fmt.Errorf("HTTP %d: %v", httpResp.StatusCode, errResp)
	}

	// Unmarshal response
	if resp != nil {
		// If wrapField is specified, wrap the response JSON into that field
		finalRespBytes := respBytes
		if wrapField != "" {
			// Create a wrapper object: {"fieldName": <original response>}
			wrapped := make(map[string]json.RawMessage)
			wrapped[wrapField] = json.RawMessage(respBytes)
			var err error
			finalRespBytes, err = json.Marshal(wrapped)
			if err != nil {
				return fmt.Errorf("wrap response: %w", err)
			}
		}

		unmarshaler := protojson.UnmarshalOptions{
			DiscardUnknown: true, // Ignore fields not in proto definition
		}
		if err := unmarshaler.Unmarshal(finalRespBytes, resp); err != nil {
			return fmt.Errorf("unmarshal response: %w (body: %s)", err, string(finalRespBytes))
		}
	}

	return nil
}
`
