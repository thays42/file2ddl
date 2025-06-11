package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DataType represents a PostgreSQL data type
type DataType struct {
	Name     string
	Priority int // Lower number means higher priority
}

var (
	// Define PostgreSQL data types in order of preference
	postgresTypes = []DataType{
		{Name: "boolean", Priority: 1},
		{Name: "smallint", Priority: 2},
		{Name: "integer", Priority: 3},
		{Name: "bigint", Priority: 4},
		{Name: "numeric", Priority: 5},
		{Name: "timestamp", Priority: 6},
		{Name: "date", Priority: 7},
		{Name: "text", Priority: 8},
	}

	// Type compatibility matrix - if a value is of type X, it can only be of type Y or less specific
	typeCompatibility = map[string][]string{
		"boolean":   {"boolean", "text"},
		"smallint":  {"smallint", "integer", "bigint", "numeric", "text"},
		"integer":   {"integer", "bigint", "numeric", "text"},
		"bigint":    {"bigint", "numeric", "text"},
		"numeric":   {"numeric", "text"},
		"timestamp": {"timestamp", "date", "text"},
		"date":      {"date", "text"},
		"text":      {"text"},
	}
)

func main() {
	// Define command line flags
	filePath := flag.String("file", "", "Path to the input file (required)")
	delimiter := flag.String("delim", "", "Field delimiter character (required)")
	flag.Parse()

	// Validate required parameters
	if *filePath == "" || *delimiter == "" {
		fmt.Println("Error: Both -file and -delim parameters are required")
		flag.Usage()
		os.Exit(1)
	}

	if len(*delimiter) != 1 {
		fmt.Println("Error: Delimiter must be a single character")
		os.Exit(1)
	}

	// Open the file
	file, err := os.Open(*filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Read the file and analyze each column
	scanner := bufio.NewScanner(file)
	var headers []string
	var columnTypes []int
	var possibleTypes [][]string
	var columnValues [][]string

	// Read headers if file is not empty
	if scanner.Scan() {
		headers = strings.Split(scanner.Text(), *delimiter)
		columnTypes = make([]int, len(headers))
		possibleTypes = make([][]string, len(headers))
		columnValues = make([][]string, len(headers))
		for i := range columnTypes {
			columnTypes[i] = 0 // Start with the most specific type (boolean)
			possibleTypes[i] = []string{"boolean", "smallint", "integer", "bigint", "numeric", "timestamp", "date", "text"}
			columnValues[i] = []string{}
		}
	}

	// Process each line
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), *delimiter)
		if len(fields) != len(headers) {
			fmt.Printf("Warning: Line has %d fields, expected %d\n", len(fields), len(headers))
			continue
		}

		// Analyze each field
		for i, field := range fields {
			fieldType := inferType(field)
			if headers[i] == "age" {
				fmt.Printf("DEBUG: value=%q, inferred_type=%s\n", field, postgresTypes[fieldType].Name)
			}
			if fieldType > columnTypes[i] {
				columnTypes[i] = fieldType
			}
			possibleTypes[i] = intersectTypes(possibleTypes[i], typeCompatibility[postgresTypes[fieldType].Name])
		}
	}

	// Print results
	fmt.Println("Column Analysis:")
	for i, header := range headers {
		fmt.Printf("%s: %s (possible types: %v)\n", header, postgresTypes[columnTypes[i]].Name, possibleTypes[i])
	}
}

// intersectTypes returns the intersection of two type lists, maintaining order
func intersectTypes(a, b []string) []string {
	result := make([]string, 0)
	bMap := make(map[string]bool)
	for _, t := range b {
		bMap[t] = true
	}
	for _, t := range a {
		if bMap[t] {
			result = append(result, t)
		}
	}
	return result
}

func inferType(value string) int {
	// Try each type in order of preference
	for i, pgType := range postgresTypes {
		switch pgType.Name {
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
	return len(postgresTypes) - 1 // Default to text
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

// contains checks if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
} 