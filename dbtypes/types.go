package dbtypes

// DataType represents a database data type
type DataType struct {
	Name     string
	Priority int // Lower number means higher priority
}

// TypeAnalyzer defines the interface for database type analysis
type TypeAnalyzer interface {
	GetTypes() []DataType
	GetTypeCompatibility() map[string][]string
}

// PostgreSQLAnalyzer implements TypeAnalyzer for PostgreSQL
type PostgreSQLAnalyzer struct{}

// GetTypes returns the PostgreSQL data types in order of preference
func (p *PostgreSQLAnalyzer) GetTypes() []DataType {
	return []DataType{
		{Name: "boolean", Priority: 1},
		{Name: "smallint", Priority: 2},
		{Name: "integer", Priority: 3},
		{Name: "bigint", Priority: 4},
		{Name: "numeric", Priority: 5},
		{Name: "timestamp", Priority: 6},
		{Name: "date", Priority: 7},
		{Name: "text", Priority: 8},
	}
}

// GetTypeCompatibility returns the PostgreSQL type compatibility matrix
func (p *PostgreSQLAnalyzer) GetTypeCompatibility() map[string][]string {
	return map[string][]string{
		"boolean":   {"boolean", "text"},
		"smallint":  {"smallint", "integer", "bigint", "numeric", "text"},
		"integer":   {"integer", "bigint", "numeric", "text"},
		"bigint":    {"bigint", "numeric", "text"},
		"numeric":   {"numeric", "text"},
		"timestamp": {"timestamp", "date", "text"},
		"date":      {"date", "text"},
		"text":      {"text"},
	}
}
