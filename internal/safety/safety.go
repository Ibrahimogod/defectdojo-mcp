// Package safety gates which DefectDojo operations the generic dispatch
// tool (dojo_call_operation) may invoke.
package safety

import "fmt"

// AllowMethod reports whether dojo_call_operation may invoke method against
// the DefectDojo API. Every method is allowed except DELETE, which needs
// enableDestructive set (DOJO_MCP_ENABLE_DESTRUCTIVE=true) — matching the
// curated tools, none of which ever issue a DELETE regardless of this flag.
func AllowMethod(method string, enableDestructive bool) error {
	if method == "DELETE" && !enableDestructive {
		return fmt.Errorf("DELETE not allowed through dojo_call_operation unless DOJO_MCP_ENABLE_DESTRUCTIVE=true")
	}
	return nil
}
