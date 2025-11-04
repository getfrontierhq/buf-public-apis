// Package main contains Go templates for code generation.
//
// These templates generate service wrapper files from proto service definitions.
package main

// serviceFileTemplate generates a service wrapper file.
//
// Template data structure:
//   - ServiceName: Name of the service (e.g., "AuthService")
//   - ImportSuffix: Package suffix for proto imports (e.g., "" or "/investments")
//   - HasFmt: Whether fmt package is needed (for path parameters)
//   - Methods: Array of MethodTemplateData
//
// Each method generates a function that:
// 1. Creates response proto
// 2. Builds the path (with or without parameters)
// 3. Calls the appropriate HTTP method (Get/Post)
// 4. Returns response and error
const serviceFileTemplate = `package client

import (
	"context"
	{{if .HasFmt}}"fmt"
	{{end}}
	"{{.HTTPClientPkg}}"
	pb "{{.ProtoPackage}}"
)

// {{.InterfaceName}} defines the interface for {{.ServiceName}}
type {{.InterfaceName}} interface {
{{range .Methods}}	// {{.Name}} makes a {{.HTTP.Method}} request to {{.HTTP.Path}}
	{{.Name}}(ctx context.Context, req *pb.{{.InputType}}) (*pb.{{.OutputType}}, error)
{{end}}}

// {{.ImplName}} provides {{.ServiceName}} operations
type {{.ImplName}} struct {
	{{.InterfaceName}}
	client *http.HTTPClient
}

{{range .Methods}}
// {{.Name}} makes a {{.HTTP.Method}} request to {{.HTTP.Path}}
func (s *{{$.ImplName}}) {{.Name}}(ctx context.Context, req *pb.{{.InputType}}) (*pb.{{.OutputType}}, error) {
	resp := &pb.{{.OutputType}}{}
	{{if .PathConstruction}}path := {{.PathConstruction}}
	{{else}}path := "{{.HTTP.Path}}"
	{{end}}{{if eq .HTTP.Method "POST"}}err := s.client.Post(ctx, path, req, resp)
	{{else}}err := s.client.Get(ctx, path, resp)
	{{end}}return resp, err
}
{{end}}`

// nestedServicesFileTemplate generates a file with a grouping client and multiple services.
//
// Template data structure:
//   - Category: Category name (e.g., "Investments")
//   - CategoryLower: Lowercase category (e.g., "investments")
//   - ImportSuffix: Package suffix for proto imports (e.g., "/investments")
//   - HasFmt: Whether fmt package is needed (for path parameters)
//   - Services: Array of ServiceTemplateData (one for each service in the category)
const nestedServicesFileTemplate = `package client

import (
	"context"
	{{if .HasFmt}}"fmt"
	{{end}}
	"{{.HTTPClientPkg}}"
	pb "{{.ProtoPackage}}"
)

// {{.InterfaceName}} defines the interface for {{.Category}} services
type {{.InterfaceName}} interface {
{{range .Services}}	Get{{.ServiceName}}() {{.InterfaceName}}
{{end}}}

// {{.ImplName}} groups {{.CategoryLower}} services
type {{.ImplName}} struct {
	{{.InterfaceName}}
{{range .Services}}	{{.PrivateField}} *{{.ImplName}}
{{end}}}

{{range .Services}}
// Get{{.ServiceName}} returns the {{.ServiceName}}
func (c *{{$.ImplName}}) Get{{.ServiceName}}() {{.InterfaceName}} {
	return c.{{.PrivateField}}
}
{{end}}

{{range $svc := .Services}}
// {{$svc.InterfaceName}} defines the interface for {{$svc.ServiceName}}
type {{$svc.InterfaceName}} interface {
{{range $svc.Methods}}	// {{.Name}} makes a {{.HTTP.Method}} request to {{.HTTP.Path}}
	{{.Name}}(ctx context.Context, req *pb.{{.InputType}}) (*pb.{{.OutputType}}, error)
{{end}}}

// {{$svc.ImplName}} provides {{$svc.ServiceName}} operations
type {{$svc.ImplName}} struct {
	{{$svc.InterfaceName}}
	client *http.HTTPClient
}

{{range $svc.Methods}}
// {{.Name}} makes a {{.HTTP.Method}} request to {{.HTTP.Path}}
func (s *{{$svc.ImplName}}) {{.Name}}(ctx context.Context, req *pb.{{.InputType}}) (*pb.{{.OutputType}}, error) {
	resp := &pb.{{.OutputType}}{}
	{{if .PathConstruction}}path := {{.PathConstruction}}
	{{else}}path := "{{.HTTP.Path}}"
	{{end}}{{if eq .HTTP.Method "POST"}}err := s.client.Post(ctx, path, req, resp)
	{{else}}err := s.client.Get(ctx, path, resp)
	{{end}}return resp, err
}
{{end}}
{{end}}`

// rootClientTemplate generates the root client file that wires everything together.
//
// Template data structure:
//   - ClientName: Name of the root client (e.g., "IniciadorClient")
//   - TopLevelServices: Array of ServiceInfo for top-level services
//   - NestedClients: Array of NestedClientInfo for nested service groups
//
// The generated client follows an immutable pattern - to change the token,
// create a new client instance rather than mutating the existing one.
// Note: Root client template doesn't need HTTPClientPkg since it's passed directly to generateRootClient
// and must be added dynamically based on configuration
const rootClientTemplate = `package client

import (
	"net/http"
	"time"

	httpclient "{{.HTTPClientPkg}}"
)

// {{.InterfaceName}} defines the interface for the root HTTP client
type {{.InterfaceName}} interface {
{{range .TopLevelServices}}	Get{{.FieldName}}() {{.InterfaceName}}
{{end}}{{range .NestedClients}}	Get{{.FieldName}}() {{.InterfaceName}}
{{end}}}

// {{.ImplName}} is the root HTTP client implementation
type {{.ImplName}} struct {
	{{.InterfaceName}}
	httpClient *httpclient.HTTPClient
{{range .TopLevelServices}}	{{.PrivateField}} *{{.ImplName}}
{{end}}{{range .NestedClients}}	{{.PrivateField}} *{{.ImplName}}
{{end}}}

{{range .TopLevelServices}}
// Get{{.FieldName}} returns the {{.TypeName}}
func (c *{{$.ImplName}}) Get{{.FieldName}}() {{.InterfaceName}} {
	return c.{{.PrivateField}}
}
{{end}}
{{range .NestedClients}}
// Get{{.FieldName}} returns the {{.TypeName}}
func (c *{{$.ImplName}}) Get{{.FieldName}}() {{.InterfaceName}} {
	return c.{{.PrivateField}}
}
{{end}}

// New{{.ClientName}} creates a new HTTP client
//
// Parameters:
//   - baseURL: API base URL (e.g., "https://data.sandbox.iniciador.com.br")
//   - token: Bearer token (empty string for unauthenticated client)
//
// The client is immutable - to change the token, create a new client instance.
func New{{.ClientName}}(baseURL, token string) *{{.ImplName}} {
	httpClient := &httpclient.HTTPClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Token:      token,
	}

	return &{{.ImplName}}{
		httpClient: httpClient,
{{range .TopLevelServices}}		{{.PrivateField}}: &{{.ImplName}}{client: httpClient},
{{end}}{{range .NestedClients}}		{{.PrivateField}}: &{{.ImplName}}{
{{range .Services}}			{{.PrivateField}}: &{{.ImplName}}{client: httpClient},
{{end}}		},
{{end}}	}
}
`
