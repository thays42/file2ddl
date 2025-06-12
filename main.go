package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"file2ddl/dbtypes"
)

// DataType represents a PostgreSQL data type
type DataType struct {
	Name     string
	Priority int // Lower number means higher priority
}

// getAnalyzer returns the appropriate TypeAnalyzer based on the database flavor
func getAnalyzer(flavor string) (dbtypes.TypeAnalyzer, error) {
	switch strings.ToLower(flavor) {
	case "postgresql":
		return &dbtypes.PostgreSQLAnalyzer{}, nil
	default:
		return nil, fmt.Errorf("unsupported database flavor: %s. Supported flavors: postgresql", flavor)
	}
}

func main() {
	// Define command line flags
	delimiter := flag.String("delim", "", "Field delimiter character (required)")
	flavor := flag.String("flavor", "postgresql", "Database flavor (default: postgresql)")
	flag.Parse()

	// Get positional arguments
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Error: File path is required as a positional argument")
		fmt.Println("Usage: file2ddl <file> -delim <delimiter>")
		os.Exit(1)
	}
	filePath := args[0]

	// Validate required parameters
	if *delimiter == "" {
		fmt.Println("Error: -delim parameter is required")
		flag.Usage()
		os.Exit(1)
	}

	if len(*delimiter) != 1 {
		fmt.Println("Error: Delimiter must be a single character")
		os.Exit(1)
	}

	// Get the appropriate analyzer
	analyzer, err := getAnalyzer(*flavor)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	headers, columnTypes := analyzeFileTypes(file, *delimiter, analyzer)

	// Print results
	fmt.Println("Column Analysis:")
	for i, header := range headers {
		fmt.Printf("%s: %s\n", header, analyzer.GetTypes()[columnTypes[i]].Name)
	}
}

// analyzeFileTypes reads the file and analyzes the types of each column
func analyzeFileTypes(file *os.File, delimiter string, analyzer dbtypes.TypeAnalyzer) ([]string, []int) {
	scanner := bufio.NewScanner(file)
	var headers []string
	var columnTypes []int

	// Read headers if file is not empty
	if scanner.Scan() {
		headers = strings.Split(scanner.Text(), delimiter)
		columnTypes = make([]int, len(headers))
		for i := range columnTypes {
			columnTypes[i] = 0 // Start with the most specific type (boolean)
		}
	}

	// Process each line
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), delimiter)
		if len(fields) != len(headers) {
			fmt.Printf("Warning: Line has %d fields, expected %d\n", len(fields), len(headers))
			continue
		}

		// Analyze each field
		for i, field := range fields {
			fieldType := inferType(field, analyzer)
			if fieldType > columnTypes[i] {
				columnTypes[i] = fieldType
				fmt.Printf("DEBUG: field %s promoted to type %s\n", headers[i], analyzer.GetTypes()[fieldType].Name)
			}
		}
	}

	return headers, columnTypes
}

func inferType(value string, analyzer dbtypes.TypeAnalyzer) int {
	// Try each type in order of preference
	types := analyzer.GetTypes()
	for i, dbType := range types {
		switch dbType.Name {
		case "boolean":
			if isBoolean(value) {
				return i
			}
		case "smallint":
			if isSmallInt(value) {
				return i
			}
		case "integer":
			if isInteger(value) {
				// If it's an integer but not a smallint, it must be an integer
				if !isSmallInt(value) {
					return i
				}
			}
		case "bigint":
			if isBigInt(value) {
				// If it's a bigint but not an integer, it must be a bigint
				if !isInteger(value) {
					return i
				}
			}
		case "numeric":
			if isNumeric(value) {
				// If it's numeric but not a bigint, it must be numeric
				if !isBigInt(value) {
					return i
				}
			}
		case "timestamp":
			if isTimestamp(value) {
				return i
			}
		case "date":
			if isDate(value) {
				return i
			}
		case "text":
			return i // text is always valid
		}
	}
	return len(types) - 1 // Default to text
}

func isBoolean(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	// Only consider explicit boolean values, not numeric 1/0
	return value == "true" || value == "false" || value == "t" || value == "f"
}

func isSmallInt(value string) bool {
	num, err := strconv.ParseInt(value, 10, 16)
	return err == nil && num >= -32768 && num <= 32767
}

func isInteger(value string) bool {
	_, err := strconv.ParseInt(value, 10, 32)
	return err == nil
}

func isBigInt(value string) bool {
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
}

func isNumeric(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

func isTimestamp(value string) bool {
	// Try common timestamp formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02T15:04:05.000",
		time.RFC3339,
	}

	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return true
		}
	}
	return false
}

func isDate(value string) bool {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return true
		}
	}
	return false
}
