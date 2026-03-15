package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// render formats a map payload into the desired output format and writes it to stdout
func render(payload any, format string) error {
	return renderTo(os.Stdout, payload, format)
}

func renderTo(w io.Writer, payload any, format string) error {
	switch format {
	case "table":
		return printTable(w, payload)
	case "json":
		fallthrough
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
