# Corsarr documentation

This directory separates released CLI behavior, the in-development Desktop
surface, and the accepted direction that has not been delivered yet.

## Current product

- [CLI architecture](ARCHITECTURE.md) describes the existing Go CLI and Docker
  Compose generator.
- [Troubleshooting](TROUBLESHOOTING.md) covers current operational failures.

## Accepted product direction

- [RFC 0001: Corsarr Desktop](rfcs/0001-corsarr-desktop.md) governs the product
  scope, user experience, runtime boundary, provisioning, security, operations,
  phased delivery, and validation plan.
- [ADR 0001: Wails v2](decisions/0001-use-wails-v2-for-desktop.md) records the
  accepted desktop framework.
- [ADR 0002: Runtime strategy](decisions/0002-docker-first-runtime-with-podman-gate.md)
  records the Docker-first MVP and the evidence gate for promoting Podman.

Phase 1 provides the reusable application catalog and Wails v2 shell. Early
Phase 2 work adds read-only Docker diagnostics and native storage-folder
inspection. The current slice persists the reviewed folder and application
selection in the operating system's user configuration directory and creates
the Corsarr-owned storage layout on explicit request. The latest development
slice additionally accepts versioned runtime consent and can install selected
digest-pinned containers when Docker is already healthy. Runtime onboarding and
Jellyseerr provisioning are not yet implemented. Current application
provisioning covers Arr root folders, qBittorrent credentials, paths and
categories, Arr download clients, Prowlarr connections, and Bazarr's Radarr and
Sonarr connections. Jellyfin provisioning covers its local administrator,
safe network default, and movie, TV, and music libraries. The RFC must not be
read as a description of the current release: only milestones explicitly
recorded as implemented have been delivered.
