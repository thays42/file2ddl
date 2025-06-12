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
	quotes := flag.String("quotes", "none", "Quote character type: none, single, or double (default: none)")
	ncols := flag.Int("ncols", 0, "Expected number of columns (optional)")

	// Parse flags after getting the file path
	flag.Parse()

	// Get positional arguments first
	if len(flag.Args()) == 0 {
		fmt.Println("Error: File path is required as a positional argument")
		fmt.Println("Usage: file2ddl <file> -delim <delimiter> [-quotes none|single|double] [-ncols <number>]")
		os.Exit(1)
	}
	filePath := flag.Args()[0]

	// Debug print for CLI parsing
	fmt.Printf("DEBUG: filePath=%q, delim=%q, quotes=%q, ncols=%d, args=%v\n", filePath, *delimiter, *quotes, *ncols, flag.Args())

	// Validate required parameters
	if *delimiter == "" {
		fmt.Println("Error: -delim parameter is required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate ncols parameter if provided
	if *ncols < 0 {
		fmt.Println("Error: ncols must be a positive integer")
		os.Exit(1)
	}

	// Extract the first character of the delimiter string
	delimChar := string((*delimiter)[0])

	// Validate quotes parameter
	if *quotes != "none" && *quotes != "single" && *quotes != "double" {
		fmt.Println("Error: quotes must be one of: none, single, double")
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

	headers, columnTypes, maxLengths, err := analyzeFileTypes(file, delimChar, *quotes, *ncols, analyzer)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println("Column Analysis:")
	for i, header := range headers {
		typeName := analyzer.GetTypes()[columnTypes[i]].Name
		if typeName == "varchar" {
			fmt.Printf("%s: varchar(%d)\n", header, maxLengths[i])
		} else {
			fmt.Printf("%s: %s\n", header, typeName)
		}
	}
}

// splitFields splits a line into fields, handling quoted fields
func splitFields(line, delim, quotes string) []string {
	if quotes == "none" {
		return strings.Split(line, delim)
	}

	var fields []string
	var current strings.Builder
	var inQuote bool
	var quoteChar rune

	if quotes == "double" {
		quoteChar = '"'
	} else {
		quoteChar = '\''
	}

	for i := 0; i < len(line); i++ {
		r := rune(line[i])

		if r == quoteChar {
			if !inQuote {
				// Start of quoted field
				inQuote = true
			} else {
				// End of quoted field
				inQuote = false
			}
			continue
		}

		if r == rune(delim[0]) && !inQuote {
			fields = append(fields, current.String())
			current.Reset()
			continue
		}

		current.WriteRune(r)
	}

	// Add the last field
	fields = append(fields, current.String())
	return fields
}

// analyzeFileTypes reads the file and analyzes the types of each column
func analyzeFileTypes(file *os.File, delimiter, quotes string, expectedCols int, analyzer dbtypes.TypeAnalyzer) ([]string, []int, []int, error) {
	scanner := bufio.NewScanner(file)
	var headers []string
	var columnTypes []int
	var maxLengths []int
	lineNum := 0

	// Read headers if file is not empty
	if scanner.Scan() {
		lineNum++
		headers = splitFields(scanner.Text(), delimiter, quotes)
		columnTypes = make([]int, len(headers))
		maxLengths = make([]int, len(headers))
		for i := range columnTypes {
			columnTypes[i] = 0 // Start with the most specific type (boolean)
			maxLengths[i] = 0
		}

		// If ncols was specified, validate header count
		if expectedCols > 0 && len(headers) != expectedCols {
			return nil, nil, nil, fmt.Errorf("header line has %d fields, expected %d", len(headers), expectedCols)
		}
	}

	// Process each line
	for scanner.Scan() {
		lineNum++
		fields := splitFields(scanner.Text(), delimiter, quotes)

		// Validate field count
		if len(fields) != len(headers) {
			return nil, nil, nil, fmt.Errorf("line %d has %d fields, expected %d", lineNum, len(fields), len(headers))
		}

		// Analyze each field
		for i, field := range fields {
			fieldType := inferType(field, analyzer)
			if fieldType > columnTypes[i] {
				columnTypes[i] = fieldType
				fmt.Printf("DEBUG: field %s promoted to type %s\n", headers[i], analyzer.GetTypes()[fieldType].Name)
			}
			if analyzer.GetTypes()[fieldType].Name == "varchar" {
				if len(field) > maxLengths[i] {
					maxLengths[i] = len(field)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("error reading file: %v", err)
	}

	return headers, columnTypes, maxLengths, nil
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
		case "varchar":
			if isVarchar(value) {
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

func isVarchar(value string) bool {
	return len(value) <= 64000
}
