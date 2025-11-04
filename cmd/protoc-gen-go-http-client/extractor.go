package main

import (
	"sort"
	"strings"

	pgs "github.com/lyft/protoc-gen-star/v2"
)

// extractServices finds all services under the specified root package.
// Only services whose package starts with rootPackage are included.
func extractServices(files []pgs.File, rootPackage string) []Service {
	var services []Service

	for _, file := range files {
		pkg := file.Package().ProtoName().String()

		// Only include services under the root package
		if !strings.HasPrefix(pkg, rootPackage) {
			continue
		}

		// Extract all services from this file
		for _, svc := range file.Services() {
			service := Service{
				Name:      svc.Name().String(),
				Package:   pkg,
				GoPackage: file.InputPath().Base(),
			}

			// Extract all methods from this service
			for _, method := range svc.Methods() {
				m := Method{
					Name:       method.Name().String(),
					InputType:  method.Input().Name().String(),
					OutputType: method.Output().Name().String(),
				}

				// Extract HTTP annotation if present
				httpInfo, err := extractHTTPInfo(method)
				if err == nil {
					m.HTTP = httpInfo
				}
				// Note: It's OK if HTTP info is missing - not all methods have HTTP annotations

				service.Methods = append(service.Methods, m)
			}

			// Sort methods by name for deterministic output
			sort.Slice(service.Methods, func(i, j int) bool {
				return service.Methods[i].Name < service.Methods[j].Name
			})

			services = append(services, service)
		}
	}

	// Sort services by name for deterministic output
	sort.Slice(services, func(i, j int) bool {
		if services[i].Package == services[j].Package {
			return services[i].Name < services[j].Name
		}
		return services[i].Package < services[j].Package
	})

	return services
}
