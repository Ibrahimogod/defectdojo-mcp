package mcptools

import (
	"net/url"
	"strconv"
)

// dojoPage is the standard DRF pagination envelope DefectDojo's list
// endpoints (findings, products, engagements, tests) all use.
type dojoPage struct {
	Count    int              `json:"count"`
	Next     string           `json:"next"`
	Previous string           `json:"previous"`
	Results  []map[string]any `json:"results"`
}

// pick trims full down to keys, keeping only the ones actually present.
func pick(full map[string]any, keys []string) map[string]any {
	out := make(map[string]any, len(keys))
	for _, k := range keys {
		if v, ok := full[k]; ok {
			out[k] = v
		}
	}
	return out
}

var findingSummaryFields = []string{
	"id", "title", "severity", "active", "verified", "false_p", "duplicate",
	"out_of_scope", "risk_accepted", "is_mitigated", "cwe", "cvssv3_score",
	"test", "found_by", "date", "age", "sla_expiration_date",
	"sla_days_remaining", "mitigated_by", "tags",
}

func summarizeFinding(full map[string]any) map[string]any {
	return pick(full, findingSummaryFields)
}

var productSummaryFields = []string{
	"id", "name", "prod_type", "business_criticality", "platform",
	"findings_count", "tags",
}

func summarizeProduct(full map[string]any) map[string]any {
	return pick(full, productSummaryFields)
}

var engagementSummaryFields = []string{
	"id", "name", "product", "status", "active", "target_start",
	"target_end", "tags",
}

func summarizeEngagement(full map[string]any) map[string]any {
	return pick(full, engagementSummaryFields)
}

var testSummaryFields = []string{
	"id", "title", "test_type_name", "scan_type", "engagement",
	"target_start", "target_end", "percent_complete", "tags",
}

func summarizeTest(full map[string]any) map[string]any {
	return pick(full, testSummaryFields)
}

// normalizeLimit caps limit to a sane default and hard maximum, matching
// DESIGN.md's pagination-safety requirement: no list tool ever lets a
// caller pull an unbounded number of results in one call.
func normalizeLimit(limit int) int {
	const (
		defaultLimit = 25
		maxLimit     = 100
	)
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func setBoolParam(q url.Values, key string, v *bool) {
	if v != nil {
		q.Set(key, strconv.FormatBool(*v))
	}
}

func setIntParam(q url.Values, key string, v int) {
	if v != 0 {
		q.Set(key, strconv.Itoa(v))
	}
}

func setStringParam(q url.Values, key, v string) {
	if v != "" {
		q.Set(key, v)
	}
}
