// Package safety gates which DefectDojo operations the generic dispatch
// tool (dojo_call_operation) may invoke.
package safety

import "fmt"

// AllowMethod reports whether dojo_call_operation may invoke method
// against the DefectDojo API. Only GET is allowed for now: write support,
// and the DOJO_MCP_ENABLE_DESTRUCTIVE-gated allowance for DELETE
// specifically, land alongside the curated write tools in a later phase.
func AllowMethod(method string) error {
	if method != "GET" {
		return fmt.Errorf("method %s not allowed yet through dojo_call_operation: only GET operations are supported until the write path ships", method)
	}
	return nil
}
