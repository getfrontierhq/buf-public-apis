package main

import (
	"testing"
)

func TestGenerateInterfaceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"AccountsService", "AccountsService"},
		{"IniciadorClient", "IniciadorClient"},
		{"TreasureTitlesService", "TreasureTitlesService"},
		{"InvestmentsClient", "InvestmentsClient"},
		{"AuthService", "AuthService"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := generateInterfaceName(tt.input)
			if result != tt.expected {
				t.Errorf("generateInterfaceName(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateImplName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"AccountsService", "AccountsServiceImpl"},
		{"IniciadorClient", "IniciadorClientImpl"},
		{"TreasureTitlesService", "TreasureTitlesServiceImpl"},
		{"InvestmentsClient", "InvestmentsClientImpl"},
		{"AuthService", "AuthServiceImpl"},
		{"", "Impl"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := generateImplName(tt.input)
			if result != tt.expected {
				t.Errorf("generateImplName(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestGeneratePrivateFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"AccountsService", "accounts"},
		{"IniciadorClient", "iniciador"},
		{"TreasureTitlesService", "treasureTitles"},
		{"InvestmentsClient", "investments"},
		{"AuthService", "auth"},
		{"LinksService", "links"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := generatePrivateFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("generatePrivateFieldName(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}
