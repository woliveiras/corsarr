# Security policy

## Supported versions

Security fixes are provided for the latest published Corsarr release. Upgrade
to the newest version before reporting a problem that may already be fixed.

| Version | Supported |
| --- | --- |
| Latest release | Yes |
| Older releases and development builds | No |

## Report a vulnerability privately

Do not open a public issue for a suspected vulnerability, exposed credential,
or report containing private data. Use
[GitHub private vulnerability reporting](https://github.com/woliveiras/corsarr/security/advisories/new)
instead. Include the affected Corsarr version and operating system, impact,
reproduction steps, and the smallest safe proof of concept.

Do not include real passwords, API keys, cookies, private tracker credentials,
download history, media filenames, or a complete home-directory path. Replace
them with clear placeholders. The project will acknowledge the report through
the private advisory and coordinate validation, remediation, and disclosure
there; no fixed response deadline is promised.

## Security boundaries

Corsarr Desktop is a local privileged orchestrator. A compromise of its backend
or the local container-runtime socket can act with the current user's runtime
authority. The socket is never exposed to the frontend, and administrative
media applications bind to localhost by default. Only Jellyfin can be exposed
to the local network through an explicit setting.

Corsarr downloads only catalog-approved immutable container-image references,
manages only labeled resources it owns, and keeps destructive application-data
removal separate from container removal. Docker Desktop, the operating system,
media applications, indexers, and user-supplied content retain their own
security and licensing responsibilities.

Diagnostic export excludes application logs and runtime sockets and redacts
credential-shaped values. Redaction is defense in depth, not a guarantee:
always inspect a report before sharing it.
