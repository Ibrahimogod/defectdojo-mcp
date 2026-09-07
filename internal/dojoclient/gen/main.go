// Command sanitize-schema strips literal null entries from every "enum"
// array in a copy of the pinned DefectDojo OpenAPI document, working around
// https://github.com/oapi-codegen/oapi-codegen/issues/1736: a null entry in
// an enum makes oapi-codegen emit the invalid Go literal <nil> instead of
// the nil keyword. DefectDojo marks nullable choice fields with a separate
// "nullable": true, so dropping the null entry from the enum list loses no
// information oapi-codegen would have used correctly anyway. The pinned
// schema file itself is never modified; this writes a sanitized copy for
// oapi-codegen to read instead.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: sanitize-schema <in.json> <out.json>")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	stripNullEnums(doc)

	out, err := json.Marshal(doc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := os.WriteFile(os.Args[2], out, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func stripNullEnums(v any) {
	switch node := v.(type) {
	case map[string]any:
		if enumVal, ok := node["enum"]; ok {
			if arr, ok := enumVal.([]any); ok {
				filtered := arr[:0]
				for _, e := range arr {
					if e != nil {
						filtered = append(filtered, e)
					}
				}
				node["enum"] = filtered
			}
		}
		for _, child := range node {
			stripNullEnums(child)
		}
	case []any:
		for _, child := range node {
			stripNullEnums(child)
		}
	}
}
