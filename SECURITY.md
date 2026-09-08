# Security Policy

## Reporting a vulnerability

Please report security vulnerabilities privately rather than opening a
public issue: use
[GitHub's private vulnerability reporting](https://github.com/ibrahimogod/defectdojo-mcp/security/advisories/new)
for this repository.

Include what you'd normally include in a report: what's affected, how to
reproduce it, and its impact. Response times aren't guaranteed (this is a
personal project), but reports are read and taken seriously.

## Scope

This server holds no DefectDojo credential of its own; every call uses
either the caller's own token (HTTP mode) or a token supplied via
environment variable at container start (stdio mode); see
[docs/DESIGN.md §7](docs/DESIGN.md#7-auth-model-detail). Vulnerabilities in
how it handles or could leak that credential (logging, error messages,
request construction) are very much in scope. Vulnerabilities in DefectDojo
itself belong in [DefectDojo's own repo](https://github.com/DefectDojo/django-DefectDojo).
