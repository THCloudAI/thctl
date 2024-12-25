// Copyright (c) 2024 THCloud.AI
// Author: OC
// Last Updated: 2024-12-25
// Description: Output formatting utilities.

package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

// Format represents the output format
type Format string

const (
	// FormatJSON outputs in JSON format
	FormatJSON Format = "json"
	// FormatYAML outputs in YAML format
	FormatYAML Format = "yaml"
	// FormatTable outputs in table format
	FormatTable Format = "table"
)

// IsValid checks if the format is valid
func (f Format) IsValid() bool {
	switch f {
	case FormatJSON, FormatYAML, FormatTable:
		return true
	default:
		return false
	}
}

// String returns the string representation of the format
func (f Format) String() string {
	return string(f)
}

// Printer handles output formatting
type Printer struct {
	format Format
}

// NewPrinter creates a new printer with the specified format
func NewPrinter(format Format) *Printer {
	if !format.IsValid() {
		format = FormatTable
	}
	return &Printer{format: format}
}

// Print formats and prints the data according to the specified format
func (p *Printer) Print(data interface{}) error {
	switch p.format {
	case FormatJSON:
		return p.printJSON(data)
	case FormatYAML:
		return p.printYAML(data)
	case FormatTable:
		return p.printTable(data)
	default:
		return fmt.Errorf("unsupported format: %s", p.format)
	}
}

func (p *Printer) printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (p *Printer) printYAML(data interface{}) error {
	return yaml.NewEncoder(os.Stdout).Encode(data)
}

func (p *Printer) printTable(data interface{}) error {
	table := tablewriter.NewWriter(os.Stdout)
	
	// Convert data to table format based on its type
	switch v := data.(type) {
	case [][]string:
		if len(v) > 0 {
			table.SetHeader(v[0])
			table.AppendBulk(v[1:])
		}
	case map[string]interface{}:
		table.SetHeader([]string{"Key", "Value"})
		for key, value := range v {
			table.Append([]string{key, fmt.Sprintf("%v", value)})
		}
	default:
		return fmt.Errorf("unsupported data type for table format: %T", data)
	}

	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	table.SetRowLine(false)
	table.SetColumnSeparator(" ")
	table.SetNoWhiteSpace(true)

	table.Render()
	return nil
}
