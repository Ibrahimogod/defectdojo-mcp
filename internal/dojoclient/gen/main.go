// Command sanitize-schema fixes two DefectDojo OpenAPI quirks in a copy of
// the pinned schema before oapi-codegen sees it. Neither touches the pinned
// file itself; this only ever writes a sanitized copy for oapi-codegen to
// read instead.
//
//  1. Nullable choice fields list null alongside their real enum values.
//     oapi-codegen turns that into the invalid Go literal <nil> instead of
//     nil (oapi-codegen/oapi-codegen#1736, still open). DefectDojo marks
//     nullability separately via "nullable": true, so dropping null from
//     the enum list loses no information oapi-codegen would have used
//     correctly anyway.
//  2. Several date/date-time filter parameters (django-filter's numeric
//     date-range shortcut, e.g. "today", "past 7 days") carry a stray
//     enum: [1,2,3,4,5,6,7,null] that contradicts the field's own
//     type/format. oapi-codegen can't generate a valid Go constant for an
//     integer enum on a date-time-typed field ("invalid constant type").
//     The field stays a plain date/date-time string; only that bogus enum
//     is dropped.
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

	sanitize(doc)

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

var dateFormats = map[string]bool{"date-time": true, "date": true}

func sanitize(v any) {
	switch node := v.(type) {
	case map[string]any:
		if _, hasEnum := node["enum"]; hasEnum {
			if format, ok := node["format"].(string); ok && dateFormats[format] {
				delete(node, "enum")
			} else if arr, ok := node["enum"].([]any); ok {
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
			sanitize(child)
		}
	case []any:
		for _, child := range node {
			sanitize(child)
		}
	}
}
