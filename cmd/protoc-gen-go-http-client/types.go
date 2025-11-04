package main

import "strings"

// ClientConfig holds the parsed configuration from the client= parameter.
// Format: "package:subdir" (e.g., "vendors.iniciador:client")
type ClientConfig struct {
	RootPackage   string // Proto package prefix (e.g., "vendors.iniciador")
	OutputSubdir  string // Output subdirectory (e.g., "client")
	ClientName    string // Root client name (e.g., "IniciadorClient", derived from package)
	GoModulePath  string // Go module path (e.g., "github.com/getfrontierhq/schema/pkg/go")
	HTTPClientPkg string // Full path to HTTP client package (computed from above)
}

// Service represents a parsed proto service with its methods.
type Service struct {
	Name      string   // Service name (e.g., "AuthService")
	Package   string   // Full proto package (e.g., "vendors.iniciador")
	GoPackage string   // Go package path for imports
	Methods   []Method // Service methods
}

// Method represents a single RPC method in a service.
type Method struct {
	Name       string    // Method name (e.g., "Authenticate")
	InputType  string    // Request message type (e.g., "AuthenticateRequest")
	OutputType string    // Response message type (e.g., "AuthenticateResponse")
	HTTP       *HTTPInfo // HTTP annotation (nil if not present)
}

// HTTPInfo contains parsed google.api.http annotation data.
type HTTPInfo struct {
	Method     string   // HTTP method: "GET", "POST", "PUT", "DELETE", "PATCH"
	Path       string   // URL path template (e.g., "/v1/data/links/{id}")
	PathParams []string // Extracted path parameters (e.g., ["id", "link_id"])
}

// ServiceTemplateData holds data for generating a service file.
type ServiceTemplateData struct {
	ServiceName   string               // e.g., "AuthService"
	ImportSuffix  string               // "" for top-level, "/investments" for nested
	HasFmt        bool                 // true if any method has path parameters
	Methods       []MethodTemplateData // all methods in the service
	HTTPClientPkg string               // Full path to HTTP client package
	ProtoPackage  string               // Full path to proto package
	// Interface generation fields
	InterfaceName string // e.g., "AuthService" (interface = base name)
	ImplName      string // e.g., "AuthServiceImpl" (struct = Impl suffix)
	PrivateField  string // e.g., "auth" (lowercase for private field)
}

// MethodTemplateData holds data for generating a single method.
type MethodTemplateData struct {
	Name             string    // e.g., "Authenticate"
	InputType        string    // e.g., "AuthenticateRequest"
	OutputType       string    // e.g., "AuthenticateResponse"
	HTTP             *HTTPInfo // HTTP method, path, and parameters
	PathConstruction string    // Go code to build the path (if has parameters)
}

// NestedServicesTemplateData holds data for generating a nested services file.
type NestedServicesTemplateData struct {
	Category      string                // e.g., "Investments"
	CategoryLower string                // e.g., "investments"
	ImportSuffix  string                // e.g., "/investments"
	HasFmt        bool                  // true if any method has path parameters
	Services      []ServiceTemplateData // all services in this category
	HTTPClientPkg string                // Full path to HTTP client package
	ProtoPackage  string                // Full path to proto package (with suffix)
	// Interface generation fields
	InterfaceName string // e.g., "InvestmentsClient" (interface = base name)
	ImplName      string // e.g., "InvestmentsClientImpl" (struct = Impl suffix)
	PrivateField  string // e.g., "investments"
}

// ServiceInfo holds information about a service for the root client.
type ServiceInfo struct {
	FieldName     string // e.g., "Auth"
	TypeName      string // e.g., "AuthService" (interface name)
	ImplName      string // e.g., "AuthServiceImpl" (struct name)
	InterfaceName string // e.g., "AuthService" (same as TypeName for compatibility)
	PrivateField  string // e.g., "auth"
}

// NestedClientInfo holds information about a nested client group.
type NestedClientInfo struct {
	FieldName     string        // e.g., "Investments"
	TypeName      string        // e.g., "InvestmentsClient" (interface name)
	ImplName      string        // e.g., "InvestmentsClientImpl" (struct name)
	InterfaceName string        // e.g., "InvestmentsClient" (same as TypeName for compatibility)
	PrivateField  string        // e.g., "investments"
	Services      []ServiceInfo // Services within this client
}

// RootClientTemplateData holds data for generating the root client file.
type RootClientTemplateData struct {
	ClientName       string             // e.g., "IniciadorClient"
	HTTPClientPkg    string             // Full path to HTTP client package
	InterfaceName    string             // e.g., "IniciadorClient" (interface = base name)
	ImplName         string             // e.g., "IniciadorClientImpl" (struct = Impl suffix)
	TopLevelServices []ServiceInfo      // Top-level services (Auth, Links)
	NestedClients    []NestedClientInfo // Nested service groups (Investments)
}

// generateInterfaceName creates an interface name from a service/client name
// The interface gets the base name (no suffix)
// Examples: "AccountsService" -> "AccountsService"
//           "IniciadorClient" -> "IniciadorClient"
func generateInterfaceName(serviceName string) string {
	return serviceName
}

// generateImplName creates an implementation struct name from a service/client name
// The implementation gets "Impl" suffix
// Examples: "AccountsService" -> "AccountsServiceImpl"
//           "IniciadorClient" -> "IniciadorClientImpl"
func generateImplName(serviceName string) string {
	return serviceName + "Impl"
}

// generatePrivateFieldName creates a private field name from a service name
// Examples: "AccountsService" -> "accounts"
//           "IniciadorClient" -> "iniciador"
//           "TreasureTitlesService" -> "treasureTitles"
func generatePrivateFieldName(serviceName string) string {
	if serviceName == "" {
		return ""
	}
	// Remove "Service" or "Client" suffix
	name := serviceName
	name = strings.TrimSuffix(name, "Service")
	name = strings.TrimSuffix(name, "Client")

	// Lowercase first letter
	if len(name) > 0 {
		return strings.ToLower(name[:1]) + name[1:]
	}
	return name
}
