package main

import (
	"fmt"
	"strings"

	pgs "github.com/lyft/protoc-gen-star/v2"
)

// parseClientConfig extracts configuration from the client= parameter.
// Expected format: "package:subdir"
// Example: "vendors.iniciador:client"
//
// Optional parameter: go_module_path=github.com/org/repo/pkg/go
// If not provided, will attempt to infer from common conventions.
func parseClientConfig(params pgs.Parameters) (*ClientConfig, error) {
	clientParam := params.Str("client")
	if clientParam == "" {
		return nil, fmt.Errorf("no client configuration specified (use: client=package:subdir)")
	}

	// Split on colon: "vendors.iniciador:client"
	parts := strings.SplitN(clientParam, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid client format: %s (expected: package:subdir)", clientParam)
	}

	rootPackage := strings.TrimSpace(parts[0])
	outputSubdir := strings.TrimSpace(parts[1])

	// Derive client name from last part of package
	// "vendors.iniciador" -> "IniciadorClient"
	packageParts := strings.Split(rootPackage, ".")
	lastPart := packageParts[len(packageParts)-1]
	clientName := strings.ToUpper(lastPart[:1]) + lastPart[1:] + "Client"

	// Get Go module path (can be overridden with go_module_path parameter)
	goModulePath := params.Str("go_module_path")
	if goModulePath == "" {
		// Default assumption: github.com/getfrontierhq/schema/pkg/go
		// This should be made configurable via parameter
		goModulePath = "github.com/getfrontierhq/schema/pkg/go"
	}

	// Compute HTTP client package path
	// go_module_path + "/" + output_subdir + "/http"
	// Example: github.com/getfrontierhq/schema/pkg/go/vendors/iniciador/httpclient/client/http
	httpClientPkg := fmt.Sprintf("%s/%s/http",
		goModulePath,
		outputSubdir,
	)

	return &ClientConfig{
		RootPackage:   rootPackage,
		OutputSubdir:  outputSubdir,
		ClientName:    clientName,
		GoModulePath:  goModulePath,
		HTTPClientPkg: httpClientPkg,
	}, nil
}
