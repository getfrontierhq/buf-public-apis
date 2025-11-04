package main

import (
	"fmt"
	"regexp"
	"strings"

	pgs "github.com/lyft/protoc-gen-star/v2"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
)

// extractHTTPInfo parses the google.api.http annotation from a method.
// Returns nil if no HTTP annotation is present.
func extractHTTPInfo(method pgs.Method) (*HTTPInfo, error) {
	// Get method options
	opts := method.Descriptor().GetOptions()
	if opts == nil {
		return nil, fmt.Errorf("no options")
	}

	// Check for google.api.http extension
	if !proto.HasExtension(opts, annotations.E_Http) {
		return nil, fmt.Errorf("no http annotation")
	}

	// Get the extension
	ext := proto.GetExtension(opts, annotations.E_Http)
	httpRule, ok := ext.(*annotations.HttpRule)
	if !ok {
		return nil, fmt.Errorf("invalid http rule type")
	}

	info := &HTTPInfo{}

	// Parse HTTP method and path from the pattern
	switch pattern := httpRule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		info.Method = "GET"
		info.Path = pattern.Get
	case *annotations.HttpRule_Post:
		info.Method = "POST"
		info.Path = pattern.Post
	case *annotations.HttpRule_Put:
		info.Method = "PUT"
		info.Path = pattern.Put
	case *annotations.HttpRule_Delete:
		info.Method = "DELETE"
		info.Path = pattern.Delete
	case *annotations.HttpRule_Patch:
		info.Method = "PATCH"
		info.Path = pattern.Patch
	default:
		return nil, fmt.Errorf("unsupported HTTP method pattern")
	}

	// Extract path parameters from the path template
	info.PathParams = extractPathParams(info.Path)

	return info, nil
}

// extractPathParams finds all path parameters in a URL template.
// Examples:
//
//	"/v1/data/links/{id}" -> ["id"]
//	"/v1/data/links/{link_id}/data/investments/{investment_id}" -> ["link_id", "investment_id"]
func extractPathParams(path string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)

	params := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}

	return params
}

// snakeToPascal converts snake_case to PascalCase.
// Examples:
//
//	"link_id" -> "LinkId"
//	"investment_id" -> "InvestmentId"
func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
