# File2DDL

A command-line tool that analyzes delimited files and infers the most appropriate PostgreSQL data types for each column.

## Features

- Analyzes delimited files (CSV, TSV, etc.)
- Infers PostgreSQL data types for each column
- Supports common data types: boolean, smallint, integer, bigint, numeric, timestamp, date, and text
- Handles common date and timestamp formats
- Validates input parameters
- Validates field count consistency across all lines

## Installation

```bash
go install github.com/yourusername/file2ddl@latest
```

## Usage

```bash
file2ddl <file> -delim <delimiter> [-quotes none|single|double] [-ncols <number>]
```

### Parameters

- `<file>`: Path to the input file (required)
- `-delim`: Single character used as field delimiter (required)
- `-quotes`: Quote character type: none, single, or double (default: none)
- `-ncols`: Expected number of columns (optional)

### Example

```bash
file2ddl data.csv -delim "," -quotes double -ncols 8
```

### Output

The tool will analyze the file and output the inferred PostgreSQL data type for each column:

```
Column Analysis:
id: integer
name: text
age: smallint
created_at: timestamp
```

## Field Count Validation

The tool validates that all lines in the file have the same number of fields:

1. When `-ncols` is not specified:
   - The number of fields in the header line is used as the expected count
   - All subsequent lines must have the same number of fields
   - An error is raised if any line has a different number of fields

2. When `-ncols` is specified:
   - The header line must have exactly the specified number of fields
   - All subsequent lines must have the same number of fields
   - An error is raised if any line has a different number of fields

Error messages include:
- The expected number of fields
- The actual number of fields found
- The line number where the mismatch occurred

## Supported Data Types

The tool tries to infer the most specific data type possible, in this order:

1. boolean
2. smallint (-32768 to 32767)
3. integer
4. bigint
5. numeric
6. timestamp
7. date
8. text (fallback type)

## Notes

- The first line of the file is assumed to be the header
- The tool will always choose the most specific type that can accommodate all values in a column 