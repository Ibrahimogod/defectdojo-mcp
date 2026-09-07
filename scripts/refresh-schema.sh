#!/usr/bin/env bash
set -euo pipefail

if [ -z "${DOJO_BASE_URL:-}" ]; then
	echo "Set DOJO_BASE_URL to your DefectDojo instance, e.g.:" >&2
	echo "  DOJO_BASE_URL=https://defectdojo.example.com ./scripts/refresh-schema.sh" >&2
	exit 1
fi

SCHEMA_URL="${DOJO_BASE_URL%/}/api/v2/oa3/schema/?format=json"
PINNED="openapi/defectdojo-schema.v3.2.100.json"
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

echo "Fetching live schema from $SCHEMA_URL ..."
curl -fsS "$SCHEMA_URL" -o "$TMP"

if diff -q "$PINNED" "$TMP" >/dev/null 2>&1; then
	echo "No changes: pinned schema matches the live instance."
	exit 0
fi

echo "Live schema differs from $PINNED:"
diff -u "$PINNED" "$TMP" || true

echo
echo "Review the diff above. If the change is intentional, copy $TMP over a"
echo "new pinned filename (bump the version in the name), update every"
echo "reference to it (openapi/, internal/dojoclient, internal/registry"
echo "go:generate directives, README), then run 'make generate'."
