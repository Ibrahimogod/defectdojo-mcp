// Package dojoclient holds the oapi-codegen-generated typed REST client
// (generated.go) plus hand-written helpers on top of it (retries, timeouts,
// error mapping, pagination) added starting Phase 1.
package dojoclient

//go:generate go run ./gen ../../openapi/defectdojo-schema.v3.2.100.json schema.sanitized.json
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen.yaml schema.sanitized.json
