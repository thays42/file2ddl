package dbtypes

import "testing"

func TestPostgreSQLAnalyzer_GetTypes(t *testing.T) {
	analyzer := &PostgreSQLAnalyzer{}
	types := analyzer.GetTypes()

	// Test that we have the expected number of types
	expectedTypes := 8
	if len(types) != expectedTypes {
		t.Errorf("Expected %d types, got %d", expectedTypes, len(types))
	}

	// Test that types are in the correct order
	expectedOrder := []string{"boolean", "smallint", "integer", "bigint", "numeric", "timestamp", "date", "text"}
	for i, expected := range expectedOrder {
		if types[i].Name != expected {
			t.Errorf("Expected type %s at position %d, got %s", expected, i, types[i].Name)
		}
	}
}

func TestPostgreSQLAnalyzer_GetTypeCompatibility(t *testing.T) {
	analyzer := &PostgreSQLAnalyzer{}
	compatibility := analyzer.GetTypeCompatibility()

	// Test that we have the expected number of type mappings
	expectedMappings := 8
	if len(compatibility) != expectedMappings {
		t.Errorf("Expected %d type mappings, got %d", expectedMappings, len(compatibility))
	}

	// Test some specific compatibility rules
	testCases := []struct {
		fromType   string
		expectedTo []string
	}{
		{"boolean", []string{"boolean", "text"}},
		{"smallint", []string{"smallint", "integer", "bigint", "numeric", "text"}},
		{"text", []string{"text"}},
	}

	for _, tc := range testCases {
		compatibleTypes, exists := compatibility[tc.fromType]
		if !exists {
			t.Errorf("Type %s not found in compatibility matrix", tc.fromType)
			continue
		}

		if len(compatibleTypes) != len(tc.expectedTo) {
			t.Errorf("Expected %d compatible types for %s, got %d", len(tc.expectedTo), tc.fromType, len(compatibleTypes))
			continue
		}

		for i, expected := range tc.expectedTo {
			if compatibleTypes[i] != expected {
				t.Errorf("Expected compatible type %s at position %d for %s, got %s", expected, i, tc.fromType, compatibleTypes[i])
			}
		}
	}
}
