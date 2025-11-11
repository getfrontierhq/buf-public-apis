// Package main implements the protoc-gen-go-http-client plugin.
// This plugin generates hierarchical HTTP clients from proto files with google.api.http annotations.
package main

import (
	"path/filepath"
	"sort"
	"strings"

	pgs "github.com/lyft/protoc-gen-star/v2"
)

// HTTPClientModule is the main plugin module.
type HTTPClientModule struct {
	*pgs.ModuleBase
}

// Name returns the name of this module.
func (m *HTTPClientModule) Name() string {
	return "http-client"
}

// Execute is called by protoc-gen-star with all proto files.
// With strategy=all, this is called once with ALL files.
func (m *HTTPClientModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	// Parse configuration
	cfg, err := parseClientConfig(m.Parameters())
	if err != nil {
		m.Fail(err.Error())
		return m.Artifacts()
	}

	m.Logf("Generating client for package: %s", cfg.RootPackage)
	m.Logf("Output subdirectory: %s", cfg.OutputSubdir)

	// Convert map to slice and sort by name for deterministic output
	var files []pgs.File
	for _, f := range targets {
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name().String() < files[j].Name().String()
	})

	// Extract services
	services := extractServices(files, cfg.RootPackage)

	m.Logf("Found %d services:", len(services))
	for _, svc := range services {
		m.Logf("  %s (%s) - %d methods", svc.Name, svc.Package, len(svc.Methods))
		for _, method := range svc.Methods {
			if method.HTTP != nil {
				m.Logf("    - %s: %s %s", method.Name, method.HTTP.Method, method.HTTP.Path)
			}
		}
	}

	// Generate base HTTP client (static file)
	httpClientPath := filepath.Join(cfg.OutputSubdir, "http", "client.gen.go")
	m.AddGeneratorFile(httpClientPath, httpClientBaseCode)
	m.Logf("Generated %s", httpClientPath)

	// Phase 4: Group services by package level
	topLevel, nested := groupServices(services, cfg.RootPackage)

	// Sort top-level services by name for deterministic output
	sort.Slice(topLevel, func(i, j int) bool {
		return topLevel[i].Name < topLevel[j].Name
	})

	// Generate top-level services (Auth, Links)
	for _, svc := range topLevel {
		code, err := generateService(svc, cfg)
		if err != nil {
			m.Logf("Error generating %s: %v", svc.Name, err)
			continue
		}

		// Service name to filename: AuthService -> auth.gen.go
		serviceName := strings.ToLower(strings.TrimSuffix(svc.Name, "Service"))
		filename := filepath.Join(cfg.OutputSubdir, serviceName+".gen.go")
		m.AddGeneratorFile(filename, code)
		m.Logf("Generated %s (%d bytes)", filename, len(code))
	}

	// Generate nested services (e.g., investments.go with InvestmentsClient)
	// Sort category names for deterministic output
	var categories []string
	for category := range nested {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	for _, category := range categories {
		categoryServices := nested[category]
		// Sort services within category for deterministic output
		sort.Slice(categoryServices, func(i, j int) bool {
			return categoryServices[i].Name < categoryServices[j].Name
		})

		code, err := generateNestedServices(category, categoryServices, cfg)
		if err != nil {
			m.Logf("Error generating %s services: %v", category, err)
			continue
		}

		filename := filepath.Join(cfg.OutputSubdir, strings.ToLower(category)+".gen.go")
		m.AddGeneratorFile(filename, code)
		m.Logf("Generated %s (%d bytes)", filename, len(code))
	}

	// Phase 5: Generate root client (client.gen.go)
	clientCode, err := generateRootClient(cfg.ClientName, topLevel, nested, cfg)
	if err != nil {
		m.Logf("Error generating root client: %v", err)
	} else {
		clientFilename := filepath.Join(cfg.OutputSubdir, "client.gen.go")
		m.AddGeneratorFile(clientFilename, clientCode)
		m.Logf("Generated %s (%d bytes)", clientFilename, len(clientCode))
	}

	return m.Artifacts()
}

// groupServices separates services into top-level and nested categories.
//
// Top-level services are those directly in the root package (e.g., vendors.iniciador).
// Nested services are grouped by their sub-package (e.g., vendors.iniciador.investments).
//
// Returns:
//   - topLevel: Services in the root package
//   - nested: Map of category -> services (e.g., "investments" -> [TreasureTitlesService, ...])
func groupServices(services []Service, rootPkg string) (topLevel []Service, nested map[string][]Service) {
	nested = make(map[string][]Service)

	for _, svc := range services {
		if svc.Package == rootPkg {
			// Top-level service (same as root package)
			topLevel = append(topLevel, svc)
		} else {
			// Nested service - extract category from package name
			// "vendors.iniciador.investments" -> "investments"
			parts := strings.Split(svc.Package, ".")
			if len(parts) > len(strings.Split(rootPkg, ".")) {
				// Get the next level after root package
				categoryIndex := len(strings.Split(rootPkg, "."))
				category := parts[categoryIndex]
				nested[category] = append(nested[category], svc)
			}
		}
	}

	return
}

func main() {
	pgs.Init(pgs.DebugEnv("DEBUG")).
		RegisterModule(&HTTPClientModule{ModuleBase: &pgs.ModuleBase{}}).
		Render()
}
