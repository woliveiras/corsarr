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

Phase 1 implementation has started with the reusable application catalog and a
Wails v2 shell. The RFC must not be read as a description of the current release:
only milestones explicitly recorded as implemented have been delivered.
