# File2DDL

A command-line tool that analyzes delimited files and infers the most appropriate PostgreSQL data types for each column.

## Features

- Analyzes delimited files (CSV, TSV, pipe-delimited, etc.)
- Infers PostgreSQL data types for each column based on actual data content
- Supports quoted fields (single quotes, double quotes, or no quotes)
- Handles VARCHAR length optimization based on actual data
- Validates field count consistency across all lines
- Extensible architecture supporting multiple database flavors
- Type promotion system that finds the most specific type possible
- Comprehensive error reporting with line numbers

## Installation

```bash
go build -o file2ddl main.go
```

## Usage

```bash
file2ddl <file> -delim <delimiter> [-flavor postgresql] [-quotes none|single|double] [-ncols <number>] [-v]
```

### Parameters

- `<file>`: Path to the input file (required, positional argument)
- `-delim`: Single character used as field delimiter (required)
- `-flavor`: Database flavor (default: postgresql) - currently only PostgreSQL is supported
- `-quotes`: Quote character handling: none, single, or double (default: none)
- `-ncols`: Expected number of columns for validation (optional)
- `-v`: Enable verbose mode with DEBUG output (optional)

### Examples

```bash
# Basic CSV analysis
file2ddl data.csv -delim ","

# CSV with double quotes
file2ddl data.csv -delim "," -quotes double

# Tab-separated file with column count validation
file2ddl data.tsv -delim $'\t' -ncols 8

# Pipe-delimited file
file2ddl data.txt -delim "|"

# Enable verbose mode to see DEBUG output
file2ddl data.csv -delim "," -v
```

## Output

The tool analyzes each column and outputs the inferred PostgreSQL data type:

```
Column Analysis:
id: smallint
name: varchar(14)
age: integer
is_active: boolean
salary: numeric
created_at: timestamp
birth_date: date
notes: varchar(16)
```

### Verbose Mode

When the `-v` flag is used, the tool outputs additional DEBUG information showing:

- Command-line parameter parsing details
- Type promotion events when columns are upgraded to more general types

Example verbose output:
```
DEBUG: filePath="data.csv", delim=",", quotes="none", ncols=0, args=["data.csv"]
DEBUG: field name promoted to type varchar
DEBUG: field salary promoted to type numeric
Column Analysis:
id: smallint
name: varchar(14)
age: integer
is_active: boolean
salary: numeric
created_at: timestamp
birth_date: date
notes: varchar(16)
```

## Supported Data Types

The tool infers types in order of specificity (most specific first):

1. **boolean** - Recognizes: `true`, `false`, `t`, `f` (case-insensitive)
2. **smallint** - Integer values from -32,768 to 32,767
3. **integer** - 32-bit integer values
4. **bigint** - 64-bit integer values  
5. **numeric** - Decimal/floating point numbers
6. **timestamp** - Date and time values in various formats:
   - `2006-01-02 15:04:05`
   - `2006-01-02T15:04:05`
   - `2006-01-02 15:04:05.000`
   - `2006-01-02T15:04:05.000`
   - RFC3339 format
7. **date** - Date-only values:
   - `2006-01-02`
   - `01/02/2006`
   - `02/01/2006`
8. **varchar(n)** - Text up to 64,000 characters (reports actual max length found)
9. **text** - Fallback for any remaining values

## Type Promotion System

The tool uses a type promotion system where columns start as the most specific type (boolean) and get promoted to more general types as needed:

- If a column has mostly numbers but one text value, it becomes `text`
- If a column has mostly small integers but one large integer, it becomes `integer`
- VARCHAR columns report the actual maximum length found: `varchar(25)`

## Quote Handling

The tool supports three quote handling modes:

- `none` (default): No quote processing, fields split directly on delimiter
- `single`: Fields enclosed in single quotes (`'field'`)
- `double`: Fields enclosed in double quotes (`"field"`)

Quoted fields can contain the delimiter character without being split:
```csv
"Smith, John","Senior Developer","123 Main St, Suite 100"
```

## Field Count Validation

The tool ensures data consistency by validating field counts:

### Without -ncols parameter:
- Uses header line field count as the expected count
- All data lines must match the header field count
- Reports line number and field counts on mismatch

### With -ncols parameter:
- Header line must have exactly the specified number of fields
- All data lines must have the same number of fields
- Provides early validation of file structure

Example error messages:
```
Error: header line has 8 fields, expected 5
Error: line 3 has 7 fields, expected 8
```

## Architecture

The tool uses a modular architecture with pluggable database type analyzers:

- `dbtypes.TypeAnalyzer` interface for different database flavors
- `dbtypes.PostgreSQLAnalyzer` for PostgreSQL-specific type inference
- Extensible design for adding MySQL, SQLite, etc. support in the future

## Error Handling

The tool provides detailed error messages for common issues:

- Missing required parameters
- File not found or unreadable
- Inconsistent field counts with line numbers
- Invalid parameter values
- Unsupported database flavors

## Assumptions

- First line of the file contains column headers
- All lines use the same delimiter consistently
- Empty fields are treated as potential NULL values
- File encoding is UTF-8 compatible

## Testing

The tool includes comprehensive tests covering:

- Individual type inference functions
- File analysis with various data types
- Quote handling in different modes
- Field count validation
- Error conditions and edge cases

Run tests with:
```bash
go test
```
