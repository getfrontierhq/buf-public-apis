package main

import (
	"fmt"
	"strings"
)

// formatServices generates a debug text output showing all services and methods.
func formatServices(services []Service) string {
	var out strings.Builder

	out.WriteString("HTTP Client Plugin - Service Extraction Debug Output\n")
	out.WriteString("=====================================================\n\n")

	for _, svc := range services {
		fmt.Fprintf(&out, "Service: %s\n", svc.Name)
		fmt.Fprintf(&out, "Package: %s\n", svc.Package)
		fmt.Fprintf(&out, "Go Package: %s\n", svc.GoPackage)
		fmt.Fprintf(&out, "Methods:\n")

		for _, m := range svc.Methods {
			fmt.Fprintf(&out, "  - %s(%s) -> %s\n", m.Name, m.InputType, m.OutputType)

			// Show HTTP info if present
			if m.HTTP != nil {
				fmt.Fprintf(&out, "    HTTP: %s %s\n", m.HTTP.Method, m.HTTP.Path)
				if len(m.HTTP.PathParams) > 0 {
					fmt.Fprintf(&out, "    Params: %v\n", m.HTTP.PathParams)
				}
			}
		}

		fmt.Fprintf(&out, "\n")
	}

	return out.String()
}
