# Corsarr documentation

This directory separates the behavior that Corsarr ships today from the accepted
direction for Corsarr Desktop.

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

The RFC is accepted but not implemented. It must not be read as a description of
the current release until its individual milestones have been delivered and
validated.
