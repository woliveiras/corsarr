# Corsarr documentation

This directory separates released CLI behavior, the implemented but
unreleased Desktop surface, and accepted work that still has empirical or
platform gates.

## Current product

- [Architecture](ARCHITECTURE.md) describes the existing Go CLI/Compose
  generator and the implemented Desktop modules.
- [Troubleshooting](TROUBLESHOOTING.md) covers current operational failures.

## Accepted product direction

- [RFC 0001: Corsarr Desktop](rfcs/0001-corsarr-desktop.md) governs the product
  scope, user experience, runtime boundary, provisioning, security, operations,
  phased delivery, and validation plan.
- [ADR 0001: Wails v2](decisions/0001-use-wails-v2-for-desktop.md) records the
  accepted desktop framework.
- [ADR 0002: Runtime strategy](decisions/0002-docker-first-runtime-with-podman-gate.md)
  records the Docker-first MVP and the evidence gate for promoting Podman.

The Desktop is a Wails v2 application in the existing Go module. It persists a
reviewed storage folder and application selection, validates the first Mac,
obtains versioned runtime consent, prepares or starts Docker Desktop on macOS,
and installs only digest-pinned containers owned by Corsarr. Its lifecycle UI
opens, starts, stops, restarts, updates, and removes each application while
keeping container removal separate from recoverable configuration archival.
The reviewed movie/TV preset includes its supporting dependencies; installed
apps cannot be removed before their installed consumers or silently dropped
from desired state. Storage requires a fresh writable, measurable 10 GiB free
capacity check before preparation and installation, with missing hardlinks
reported as a non-blocking efficiency warning.

Provisioning currently covers Arr root folders and API keys, qBittorrent's
generated Keychain credential, paths and categories, Arr download clients,
Prowlarr connections, Bazarr's Radarr/Sonarr connections, Jellyfin's local
administrator and libraries, and Seerr's Jellyfin/Radarr/Sonarr setup. Updates
create a private configuration backup, replace the image, verify readiness, and
restore the previous container image on failure without claiming to reverse a
database migration.

The Desktop also supports explicit Jellyfin LAN access and local-address
discovery, optional macOS login recovery for existing resources, redacted
diagnostic export, localized catalogs, and catalog-generated legal credits.
Windows/Linux onboarding, native secure stores on those platforms, Apple
Developer ID signing/notarization, and the empirical Podman promotion matrix
remain open gates. This is implemented development behavior, not a statement
that a public Desktop release has shipped.

The local macOS arm64 bundle builds, self-signs, declares macOS 14+, and has
passed deterministic suites plus visual smoke checks on the first Mac. No media
stack was installed during automated development validation: runtime terms,
image downloads, and application installation remain an explicit user action.

The Desktop also includes a catalog-generated "Aplicativos e licenças" screen.
It credits every offered application and image maintainer, identifies Docker and
Podman separately, shows the installed or approved digest, and opens only
backend-allowlisted official, source, license, image, and support links.
