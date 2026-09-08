// Package dojoclient is a small hand-written REST client for the
// DefectDojo API. The curated MCP tools use its typed-ish helpers; the
// generic dispatch tools (dojo_call_operation) use Do directly against
// whatever method/path the operation registry describes, since those
// aren't known until request time.
package dojoclient
