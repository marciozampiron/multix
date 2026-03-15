package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// OutputFormat defines the requested format type
type OutputFormat string

const (
	FormatJSON  OutputFormat = "json"
	FormatTable OutputFormat = "table"
)

// Print formats a map payload into the desired output format and writes it
func Print(payload any, format OutputFormat) error {
	return PrintTo(os.Stdout, payload, format)
}

func PrintTo(w io.Writer, payload any, format OutputFormat) error {
	switch format {
	case FormatJSON:
		return printJSON(w, payload)
	case FormatTable:
		// For MVP, table fallback is simple formatted text,
		// relying on `jq` or custom text blocks.
		// A full table writer (like pterm or text/tabwriter) can be implemented here.
		return printTable(w, payload)
	default:
		return printJSON(w, payload) // Default to JSON
	}
}

func printJSON(w io.Writer, payload any) error {
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(b))
	return nil
}

func printTable(w io.Writer, payload any) error {
	// MVP: If payload is a map, iterate and print K:V cleanly
	if m, ok := payload.(map[string]any); ok {
		for k, v := range m {
			fmt.Fprintf(w, "%s:\t%v\n", k, v)
		}
		return nil
	}
	// Fallback if not a map
	fmt.Fprintf(w, "%v\n", payload)
	return nil
}
