package main

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Run tests
	os.Exit(m.Run())
}

func TestTypeInference(t *testing.T) {
	// Test cases for individual type inference functions
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"boolean_true", "true", "boolean"},
		{"boolean_false", "false", "boolean"},
		{"boolean_t", "t", "boolean"},
		{"boolean_f", "f", "boolean"},
		{"numeric_one", "1", "smallint"},
		{"numeric_zero", "0", "smallint"},
		{"smallint_min", "-32768", "smallint"},
		{"smallint_max", "32767", "smallint"},
		{"integer", "32768", "integer"},
		{"bigint", "9223372036854775807", "bigint"},
		{"numeric", "123.45", "numeric"},
		{"timestamp", "2024-03-20 10:30:00", "timestamp"},
		{"date", "2024-03-20", "date"},
		{"text", "Hello, World!", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := postgresTypes[inferType(tt.value)].Name
			if got != tt.expected {
				t.Errorf("inferType(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestTypeCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		types    []string
		expected []string
	}{
		{
			name:     "timestamp compatibility",
			types:    typeCompatibility["timestamp"],
			expected: []string{"timestamp", "date", "text"},
		},
		{
			name:     "smallint compatibility",
			types:    typeCompatibility["smallint"],
			expected: []string{"smallint", "integer", "bigint", "numeric", "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.types) != len(tt.expected) {
				t.Errorf("got %d types, want %d", len(tt.types), len(tt.expected))
				return
			}
			for i, typ := range tt.types {
				if typ != tt.expected[i] {
					t.Errorf("type[%d] = %s, want %s", i, typ, tt.expected[i])
				}
			}
		})
	}
}

func TestFileAnalysis(t *testing.T) {
	// Create a temporary file with test data
	tmpFile := "testdata/sample.csv"
	
	// Test the file analysis
	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Read and analyze the file
	scanner := bufio.NewScanner(file)
	var headers []string
	var columnTypes []int
	var possibleTypes [][]string

	// Read headers
	if scanner.Scan() {
		headers = strings.Split(scanner.Text(), ",")
		columnTypes = make([]int, len(headers))
		possibleTypes = make([][]string, len(headers))
		for i := range columnTypes {
			columnTypes[i] = len(postgresTypes) - 1 // Start with text type
			possibleTypes[i] = []string{"boolean", "smallint", "integer", "bigint", "numeric", "timestamp", "date", "text"}
		}
	}

	// Process each line
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")
		if len(fields) != len(headers) {
			t.Errorf("Line has %d fields, expected %d", len(fields), len(headers))
			continue
		}

		for i, field := range fields {
			fieldType := inferType(field)
			if fieldType < columnTypes[i] {
				columnTypes[i] = fieldType
			}
			possibleTypes[i] = intersectTypes(possibleTypes[i], typeCompatibility[postgresTypes[fieldType].Name])
		}
	}

	// Expected types for each column
	expectedTypes := map[string]string{
		"id":         "smallint",
		"name":       "text",
		"age":        "integer", // Because it contains values outside smallint range
		"is_active":  "boolean",
		"salary":     "numeric",
		"created_at": "timestamp",
		"birth_date": "date",
		"notes":      "text",
	}

	// Verify the inferred types
	for i, header := range headers {
		expected := expectedTypes[header]
		got := postgresTypes[columnTypes[i]].Name
		if header == "id" || header == "age" {
			t.Logf("DEBUG: header=%s, possible types=%v, got=%s, expected=%s", header, possibleTypes[i], got, expected)
		}
		if got != expected {
			t.Errorf("Column %s: got type %s, want %s", header, got, expected)
		}
	}
} 