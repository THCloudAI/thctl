package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Format represents the output format
type Format string

const (
	// FormatJSON represents JSON output format
	FormatJSON Format = "json"
	// FormatYAML represents YAML output format
	FormatYAML Format = "yaml"
	// FormatTable represents table output format
	FormatTable Format = "table"
)

// Print prints data in the specified format
func Print(data interface{}, format Format) error {
	switch format {
	case FormatJSON:
		return printJSON(data)
	case FormatYAML:
		return printYAML(data)
	case FormatTable:
		return printTable(data)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// printJSON prints data in JSON format
func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printYAML prints data in YAML format
func printYAML(data interface{}) error {
	return yaml.NewEncoder(os.Stdout).Encode(data)
}

// printTable prints data in table format
func printTable(data interface{}) error {
	// TODO: Implement table output format
	return fmt.Errorf("table output format not implemented yet")
}
