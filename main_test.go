package main

import (
	"bufio"
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
		{"varchar", "Hello, World!", "varchar"},
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

	// Test cases for different ncols values
	testCases := []struct {
		name        string
		ncols       int
		wantErr     bool
		errContains string
	}{
		{
			name:    "no ncols specified",
			ncols:   0,
			wantErr: false,
		},
		{
			name:    "correct ncols specified",
			ncols:   8,
			wantErr: false,
		},
		{
			name:        "incorrect ncols specified",
			ncols:       5,
			wantErr:     true,
			errContains: "header line has 8 fields, expected 5",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset file position
			file.Seek(0, 0)

			// Analyze the file using the new function
			headers, columnTypes, maxLengths, err := analyzeFileTypes(file, ",", "none", tc.ncols, analyzer)

			if tc.wantErr {
				if err == nil {
					t.Error("analyzeFileTypes() error = nil, want error")
					return
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("analyzeFileTypes() error = %v, want error containing %v", err, tc.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("analyzeFileTypes() error = %v, want nil", err)
				return
			}

			// Expected types for each column
			expectedTypes := map[string]string{
				"id":         "smallint",
				"name":       "varchar",
				"age":        "integer",
				"is_active":  "boolean",
				"salary":     "numeric",
				"created_at": "timestamp",
				"birth_date": "date",
				"notes":      "varchar",
			}

			// Expected max lengths for varchar columns (from testdata/sample.csv)
			expectedMaxLengths := map[string]int{
				"name":  14, // "Charlie Wilson"
				"notes": 16, // "Regular employee" or "Senior developer"
			}

			// Verify the inferred types and max lengths
			for i, header := range headers {
				expected := expectedTypes[header]
				got := analyzer.GetTypes()[columnTypes[i]].Name
				if got != expected {
					t.Errorf("Column %s: got type %s, want %s", header, got, expected)
				}
				if got == "varchar" {
					expectedLen := expectedMaxLengths[header]
					if maxLengths[i] != expectedLen {
						t.Errorf("Column %s: got varchar(%d), want varchar(%d)", header, maxLengths[i], expectedLen)
					}
				}
			}
		})
	}
}

func TestInvalidFieldCount(t *testing.T) {
	// Create a temporary file with invalid data
	tmpFile := "testdata/invalid_sample.csv"

	// Test the file analysis
	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Create analyzer
	analyzer := &dbtypes.PostgreSQLAnalyzer{}

	// Analyze the file
	_, _, _, err = analyzeFileTypes(file, ",", "none", 0, analyzer)
	if err == nil {
		t.Error("analyzeFileTypes() error = nil, want error")
		return
	}

	// Verify error message contains line number and field count
	if !strings.Contains(err.Error(), "line") || !strings.Contains(err.Error(), "fields") {
		t.Errorf("analyzeFileTypes() error = %v, want error containing line number and field count", err)
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

func TestQuotedFieldHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		delim    string
		quotes   string
		expected []string
	}{
		{
			name:     "unquoted fields",
			input:    "a,b,c",
			delim:    ",",
			quotes:   "none",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "double quoted fields",
			input:    `"a","b","c"`,
			delim:    ",",
			quotes:   "double",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "single quoted fields",
			input:    "'a','b','c'",
			delim:    ",",
			quotes:   "single",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "mixed quoted and unquoted",
			input:    `"a",b,"c"`,
			delim:    ",",
			quotes:   "double",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "quoted fields with delimiter inside",
			input:    `"a,b","c,d"`,
			delim:    ",",
			quotes:   "double",
			expected: []string{"a,b", "c,d"},
		},
		{
			name:     "quoted fields with spaces",
			input:    `"a b","c d"`,
			delim:    ",",
			quotes:   "double",
			expected: []string{"a b", "c d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := splitFields(tt.input, tt.delim, tt.quotes)
			if len(fields) != len(tt.expected) {
				t.Errorf("got %d fields, want %d", len(fields), len(tt.expected))
				return
			}
			for i, field := range fields {
				if field != tt.expected[i] {
					t.Errorf("field[%d] = %q, want %q", i, field, tt.expected[i])
				}
			}
		})
	}
}

func TestQuotedFileAnalysis(t *testing.T) {
	// Create a temporary file with test data
	tmpFile := "testdata/quoted_sample.csv"

	// Test the file analysis
	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Create analyzer
	analyzer := &dbtypes.PostgreSQLAnalyzer{}

	// Analyze the file using the new function
	headers, columnTypes, maxLengths, err := analyzeFileTypes(file, ",", "double", 0, analyzer)
	if err != nil {
		t.Fatalf("Failed to analyze file: %v", err)
	}

	// Expected types for each column
	expectedTypes := map[string]string{
		"id":          "smallint",
		"name":        "varchar",
		"description": "varchar",
		"address":     "varchar",
		"phone":       "varchar",
		"email":       "varchar",
		"created_at":  "timestamp",
		"notes":       "varchar",
	}

	// Expected max lengths for varchar columns (from testdata/quoted_sample.csv)
	expectedMaxLengths := map[string]int{
		"name":        13, // "Williams, Bob"
		"description": 16, // "Senior Developer"
		"address":     22, // "123 Main St, Suite 100"
		"phone":       8,  // "555-1234"
		"email":       24, // "john.smith@example.com"
		"notes":       16, // "Regular employee"
	}

	// Verify the inferred types and max lengths
	for i, header := range headers {
		expected := expectedTypes[header]
		got := analyzer.GetTypes()[columnTypes[i]].Name
		if got != expected {
			t.Errorf("Column %s: got type %s, want %s", header, got, expected)
		}
		if got == "varchar" {
			expectedLen := expectedMaxLengths[header]
			if maxLengths[i] != expectedLen {
				t.Errorf("Column %s: got varchar(%d), want varchar(%d)", header, maxLengths[i], expectedLen)
			}
		}
	}

	// Verify that quoted fields with commas are handled correctly
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		headers := splitFields(scanner.Text(), ",", "double")
		if len(headers) != 8 {
			t.Errorf("Expected 8 headers, got %d", len(headers))
		}
	}

	// Read first data line
	if scanner.Scan() {
		fields := splitFields(scanner.Text(), ",", "double")
		if len(fields) != 8 {
			t.Errorf("Expected 8 fields, got %d", len(fields))
		}
		// Verify that fields with commas are preserved
		if fields[1] != "Smith, John" {
			t.Errorf("Expected 'Smith, John', got %q", fields[1])
		}
		if fields[3] != "123 Main St, Suite 100" {
			t.Errorf("Expected '123 Main St, Suite 100', got %q", fields[3])
		}
	}
}
