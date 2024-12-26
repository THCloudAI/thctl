package output

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
)

// Format represents the output format
type Format string

const (
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
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

// Printer represents an output printer
type Printer struct {
	format Format
}

// NewPrinter creates a new printer
func NewPrinter(format Format) *Printer {
	return &Printer{format: format}
}

// Print prints the data in the specified format
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
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (p *Printer) printYAML(data interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (p *Printer) printTable(data interface{}) error {
	// For now, just print as JSON
	return p.printJSON(data)
}
