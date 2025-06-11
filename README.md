# File2DDL

A command-line tool that analyzes delimited files and infers the most appropriate PostgreSQL data types for each column.

## Features

- Analyzes delimited files (CSV, TSV, etc.)
- Infers PostgreSQL data types for each column
- Supports common data types: boolean, smallint, integer, bigint, numeric, timestamp, date, and text
- Handles common date and timestamp formats
- Validates input parameters

## Installation

```bash
go install github.com/yourusername/file2ddl@latest
```

## Usage

```bash
file2ddl -file <path-to-file> -delim <delimiter>
```

### Parameters

- `-file`: Path to the input file (required)
- `-delim`: Single character used as field delimiter (required)

### Example

```bash
file2ddl -file data.csv -delim ","
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
- If a line has a different number of fields than the header, it will be skipped with a warning
- The tool will always choose the most specific type that can accommodate all values in a column 