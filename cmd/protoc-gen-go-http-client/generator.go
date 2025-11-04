// Package main contains service code generation logic.
//
// This file transforms proto service definitions into Go service wrappers.
package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

// generateService generates Go code for a single service.
//
// This function:
// 1. Determines the import suffix based on package name
// 2. Processes each method to build path construction code
// 3. Executes the service template
// 4. Returns the generated Go code
//
// Parameters:
//   - svc: Service definition with methods and HTTP info
//   - cfg: Client configuration with paths
//
// Returns:
//   - Generated Go code as a string
//   - Error if template execution fails
func generateService(svc Service, cfg *ClientConfig) (string, error) {
	// Compute import suffix dynamically based on package difference
	// "vendors.iniciador" with root "vendors.iniciador" → ""
	// "vendors.iniciador.investments" with root "vendors.iniciador" → "/investments"
	importSuffix := computeImportSuffix(svc.Package, cfg.RootPackage)

	// Compute proto package path
	protoPackage := fmt.Sprintf("%s/%s%s",
		cfg.GoModulePath,
		strings.Replace(cfg.RootPackage, ".", "/", -1),
		importSuffix,
	)

	data := ServiceTemplateData{
		ServiceName:   svc.Name,
		ImportSuffix:  importSuffix,
		HTTPClientPkg: cfg.HTTPClientPkg,
		ProtoPackage:  protoPackage,
		InterfaceName: generateInterfaceName(svc.Name),
		ImplName:      generateImplName(svc.Name),
		PrivateField:  generatePrivateFieldName(svc.Name),
	}

	// Process methods
	for _, m := range svc.Methods {
		if m.HTTP == nil {
			// Skip methods without HTTP annotations
			continue
		}

		methodData := MethodTemplateData{
			Name:       m.Name,
			InputType:  m.InputType,
			OutputType: m.OutputType,
			HTTP:       m.HTTP,
		}

		// Build path construction code if method has path parameters
		if len(m.HTTP.PathParams) > 0 {
			data.HasFmt = true // Need fmt.Sprintf
			methodData.PathConstruction = buildPathConstruction(m.HTTP.Path, m.HTTP.PathParams)
		}

		data.Methods = append(data.Methods, methodData)
	}

	// Execute template
	tmpl, err := template.New("service").Parse(serviceFileTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// computeImportSuffix calculates the import suffix based on package difference.
//
// Examples:
//   - Package: "vendors.iniciador", Root: "vendors.iniciador" → ""
//   - Package: "vendors.iniciador.investments", Root: "vendors.iniciador" → "/investments"
//   - Package: "vendors.other.api.v1", Root: "vendors.other" → "/api/v1"
func computeImportSuffix(pkg, rootPkg string) string {
	if pkg == rootPkg {
		return ""
	}

	// Remove root package prefix
	if !strings.HasPrefix(pkg, rootPkg+".") {
		return ""
	}

	suffix := strings.TrimPrefix(pkg, rootPkg+".")
	// Convert dots to slashes: "investments.api" → "/investments/api"
	return "/" + strings.Replace(suffix, ".", "/", -1)
}

// buildPathConstruction generates Go code to construct a URL path with parameters.
//
// This function:
// 1. Replaces {param} placeholders with %s for fmt.Sprintf
// 2. Converts snake_case parameter names to PascalCase field names
// 3. Generates the fmt.Sprintf call with proper field access
//
// Examples:
//
//   - Path: "/v1/data/links/{id}", Params: ["id"]
//     Returns: `fmt.Sprintf("/v1/data/links/%s", req.Id)`
//
//   - Path: "/v1/data/links/{link_id}/investments/{investment_id}", Params: ["link_id", "investment_id"]
//     Returns: `fmt.Sprintf("/v1/data/links/%s/investments/%s", req.LinkId, req.InvestmentId)`
//
// Parameters:
//   - path: URL path template with {param} placeholders
//   - params: List of parameter names (snake_case)
//
// Returns:
//   - Go code string for fmt.Sprintf call
func buildPathConstruction(path string, params []string) string {
	// Replace {param} with %s for fmt.Sprintf template
	template := path
	for _, param := range params {
		template = strings.Replace(template, "{"+param+"}", "%s", 1)
	}

	// Build argument list: req.Param1, req.Param2, ...
	var args []string
	for _, param := range params {
		// Convert snake_case to PascalCase: link_id → LinkId
		fieldName := snakeToPascal(param)
		args = append(args, "req."+fieldName)
	}

	// Generate: fmt.Sprintf("template", args...)
	return fmt.Sprintf(`fmt.Sprintf("%s", %s)`, template, strings.Join(args, ", "))
}

// generateNestedServices generates Go code for multiple services in a category.
//
// This function creates a file with:
// 1. A grouping client (e.g., InvestmentsClient)
// 2. All service structs in the category (e.g., TreasureTitlesService, FundsService)
// 3. All methods for each service
//
// Parameters:
//   - category: Category name (e.g., "investments")
//   - services: All services in this category
//   - cfg: Client configuration with paths
//
// Returns:
//   - Generated Go code as a string
//   - Error if template execution fails
func generateNestedServices(category string, services []Service, cfg *ClientConfig) (string, error) {
	// Capitalize category for struct names: "investments" -> "Investments"
	categoryCapitalized := strings.ToUpper(category[:1]) + category[1:]

	importSuffix := "/" + category

	// Compute proto package path for nested services
	protoPackage := fmt.Sprintf("%s/%s%s",
		cfg.GoModulePath,
		strings.Replace(cfg.RootPackage, ".", "/", -1),
		importSuffix,
	)

	data := NestedServicesTemplateData{
		Category:      categoryCapitalized,
		CategoryLower: category,
		ImportSuffix:  importSuffix,
		HTTPClientPkg: cfg.HTTPClientPkg,
		ProtoPackage:  protoPackage,
		InterfaceName: generateInterfaceName(categoryCapitalized + "Client"),
		ImplName:      generateImplName(categoryCapitalized + "Client"),
		PrivateField:  generatePrivateFieldName(categoryCapitalized + "Client"),
	}

	// Process each service in the category
	for _, svc := range services {
		serviceData := ServiceTemplateData{
			ServiceName:   svc.Name,
			InterfaceName: generateInterfaceName(svc.Name),
			ImplName:      generateImplName(svc.Name),
			PrivateField:  generatePrivateFieldName(svc.Name),
		}

		// Process methods
		for _, m := range svc.Methods {
			if m.HTTP == nil {
				continue
			}

			methodData := MethodTemplateData{
				Name:       m.Name,
				InputType:  m.InputType,
				OutputType: m.OutputType,
				HTTP:       m.HTTP,
			}

			// Build path construction if method has path parameters
			if len(m.HTTP.PathParams) > 0 {
				data.HasFmt = true
				methodData.PathConstruction = buildPathConstruction(m.HTTP.Path, m.HTTP.PathParams)
			}

			serviceData.Methods = append(serviceData.Methods, methodData)
		}

		data.Services = append(data.Services, serviceData)
	}

	// Execute template
	tmpl, err := template.New("nested").Parse(nestedServicesFileTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// generateRootClient generates the root client file that wires everything together.
//
// This function creates a client.go file with:
// 1. The root client struct (e.g., IniciadorClient)
// 2. A constructor function (e.g., NewIniciadorClient)
// 3. Initialization of all services and nested clients
//
// The generated client follows an immutable pattern - to change the token,
// users must create a new client instance.
//
// Parameters:
//   - clientName: Name of the root client (e.g., "IniciadorClient")
//   - topLevel: Top-level services (e.g., AuthService, LinksService)
//   - nested: Map of category -> services (e.g., "investments" -> [TreasureTitlesService, ...])
//
// Returns:
//   - Generated Go code as a string
//   - Error if template execution fails
func generateRootClient(clientName string, topLevel []Service, nested map[string][]Service, cfg *ClientConfig) (string, error) {
	data := RootClientTemplateData{
		ClientName:    clientName,
		HTTPClientPkg: cfg.HTTPClientPkg,
		InterfaceName: generateInterfaceName(clientName),
		ImplName:      generateImplName(clientName),
	}

	// Process top-level services (already sorted in caller)
	for _, svc := range topLevel {
		// AuthService -> Auth
		fieldName := strings.TrimSuffix(svc.Name, "Service")
		data.TopLevelServices = append(data.TopLevelServices, ServiceInfo{
			FieldName:     fieldName,
			TypeName:      svc.Name,
			ImplName:      generateImplName(svc.Name),
			InterfaceName: generateInterfaceName(svc.Name),
			PrivateField:  generatePrivateFieldName(svc.Name),
		})
	}

	// Process nested clients - sort category keys for deterministic output
	var categories []string
	for category := range nested {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	for _, category := range categories {
		services := nested[category]
		// Sort services within category for deterministic output
		sort.Slice(services, func(i, j int) bool {
			return services[i].Name < services[j].Name
		})
		// "investments" -> "Investments"
		categoryCapitalized := strings.ToUpper(category[:1]) + category[1:]

		nestedClient := NestedClientInfo{
			FieldName:     categoryCapitalized,
			TypeName:      categoryCapitalized + "Client",
			ImplName:      generateImplName(categoryCapitalized + "Client"),
			InterfaceName: generateInterfaceName(categoryCapitalized + "Client"),
			PrivateField:  generatePrivateFieldName(categoryCapitalized + "Client"),
		}

		// Add all services in this category
		for _, svc := range services {
			nestedClient.Services = append(nestedClient.Services, ServiceInfo{
				FieldName:     svc.Name,
				TypeName:      svc.Name,
				ImplName:      generateImplName(svc.Name),
				InterfaceName: generateInterfaceName(svc.Name),
				PrivateField:  generatePrivateFieldName(svc.Name),
			})
		}

		data.NestedClients = append(data.NestedClients, nestedClient)
	}

	// Execute template
	tmpl, err := template.New("rootclient").Parse(rootClientTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
