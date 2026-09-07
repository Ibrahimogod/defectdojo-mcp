// Package registry describes every DefectDojo API operation (operationId,
// method, path, parameter/request/response schema) from the pinned OpenAPI
// document, so the discovery/dispatch MCP tools can list, describe, and
// invoke any of the 191 operations without hand-written wrappers.
package registry

//go:generate go run ./gen ../../openapi/defectdojo-schema.v3.2.100.json registry_gen.go

import "encoding/json"

type OperationEntry struct {
	OperationID       string
	Method            string
	Path              string
	Tag               string
	Summary           string
	Parameters        json.RawMessage
	RequestBodySchema json.RawMessage
	ResponseSchema    json.RawMessage
}

func All() []OperationEntry {
	return Operations
}

func ByOperationID(id string) (OperationEntry, bool) {
	for _, op := range Operations {
		if op.OperationID == id {
			return op, true
		}
	}
	return OperationEntry{}, false
}
