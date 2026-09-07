<#
.SYNOPSIS
    Re-fetches the DefectDojo OpenAPI schema from the live instance and diffs
    it against the pinned copy in openapi/. Windows-friendly entry point for
    `make refresh-schema` (this repo runs on Windows, where `make` may not be
    installed).
#>

$ErrorActionPreference = "Stop"

if (-not $env:DOJO_BASE_URL) {
	Write-Error "Set DOJO_BASE_URL to your DefectDojo instance, e.g.:`n  `$env:DOJO_BASE_URL = 'https://defectdojo.example.com'; .\scripts\refresh-schema.ps1"
	exit 1
}

$schemaUrl = "$($env:DOJO_BASE_URL.TrimEnd('/'))/api/v2/oa3/schema/?format=json"
$repoRoot = Split-Path -Parent $PSScriptRoot
$pinned = Join-Path $repoRoot "openapi\defectdojo-schema.v3.2.100.json"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) "defectdojo-schema-live-$([guid]::NewGuid().ToString('N')).json"

try {
	Write-Host "Fetching live schema from $schemaUrl ..."
	Invoke-WebRequest -Uri $schemaUrl -OutFile $tmp -UseBasicParsing

	$pinnedContent = Get-Content -Raw -Path $pinned
	$liveContent = Get-Content -Raw -Path $tmp

	if ($pinnedContent -eq $liveContent) {
		Write-Host "No changes: pinned schema matches the live instance."
		exit 0
	}

	Write-Host "Live schema differs from $pinned"
	$diff = Compare-Object -ReferenceObject (Get-Content $pinned) -DifferenceObject (Get-Content $tmp)
	$diff | Format-Table -AutoSize | Out-String | Write-Host

	Write-Host ""
	Write-Host "Review the diff above. If the change is intentional, copy $tmp over a"
	Write-Host "new pinned filename (bump the version in the name), update every"
	Write-Host "reference to it (openapi/, internal/dojoclient, internal/registry"
	Write-Host "go:generate directives, README), then run 'go generate ./...'."
}
finally {
	if (Test-Path $tmp) { Remove-Item $tmp -Force }
}
