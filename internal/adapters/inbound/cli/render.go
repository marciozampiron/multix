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
	// MVP: Convert payload to JSON, then to map, to uniformly print K:V cleanly
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		// Fallback if not an object
		fmt.Fprintf(w, "%v\n", payload)
		return nil
	}

	for k, v := range m {
		if v != nil && v != "" {
			fmt.Fprintf(w, "%s:\t%v\n", k, v)
		}
	}
	return nil
}
