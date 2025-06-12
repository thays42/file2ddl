package main

import (
	"os"
	"strings"
	"testing"

	"file2ddl/dbtypes"
)

func TestMain(m *testing.M) {
	// Run tests
	os.Exit(m.Run())
}

func TestTypeInference(t *testing.T) {
	analyzer := &dbtypes.PostgreSQLAnalyzer{}

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
			got := analyzer.GetTypes()[inferType(tt.value, analyzer)].Name
			if got != tt.expected {
				t.Errorf("inferType(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestTypeCompatibility(t *testing.T) {
	analyzer := &dbtypes.PostgreSQLAnalyzer{}
	compatibility := analyzer.GetTypeCompatibility()

	tests := []struct {
		name     string
		types    []string
		expected []string
	}{
		{
			name:     "timestamp compatibility",
			types:    compatibility["timestamp"],
			expected: []string{"timestamp", "date", "text"},
		},
		{
			name:     "smallint compatibility",
			types:    compatibility["smallint"],
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

	// Create analyzer
	analyzer := &dbtypes.PostgreSQLAnalyzer{}

	// Analyze the file using the new function
	headers, columnTypes := analyzeFileTypes(file, ",", analyzer)

	// Expected types for each column
	expectedTypes := map[string]string{
		"id":         "smallint",
		"name":       "text",
		"age":        "integer",
		"is_active":  "boolean",
		"salary":     "numeric",
		"created_at": "timestamp",
		"birth_date": "date",
		"notes":      "text",
	}

	// Verify the inferred types
	for i, header := range headers {
		expected := expectedTypes[header]
		got := analyzer.GetTypes()[columnTypes[i]].Name
		if got != expected {
			t.Errorf("Column %s: got type %s, want %s", header, got, expected)
		}
	}
}

func TestGetAnalyzer(t *testing.T) {
	tests := []struct {
		name        string
		flavor      string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid postgresql flavor",
			flavor:  "postgresql",
			wantErr: false,
		},
		{
			name:    "valid postgresql flavor case insensitive",
			flavor:  "PostgreSQL",
			wantErr: false,
		},
		{
			name:        "invalid flavor",
			flavor:      "mysql",
			wantErr:     true,
			errContains: "unsupported database flavor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := getAnalyzer(tt.flavor)
			if tt.wantErr {
				if err == nil {
					t.Error("getAnalyzer() error = nil, want error")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("getAnalyzer() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("getAnalyzer() error = %v, want nil", err)
				return
			}
			if analyzer == nil {
				t.Error("getAnalyzer() analyzer = nil, want non-nil")
			}
		})
	}
}
