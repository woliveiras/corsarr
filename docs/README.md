# Corsarr documentation

This directory separates released behavior, operational instructions, accepted
architecture, and work that still has empirical or platform gates.

## Current product

- [Architecture](ARCHITECTURE.md) describes the existing Go CLI/Compose
  generator and the implemented Desktop modules.
- [Troubleshooting](TROUBLESHOOTING.md) covers current operational failures.
- [Desktop acceptance](DESKTOP_ACCEPTANCE.md) separates repeatable first-Mac
  evidence from human acceptance and external release gates.
- [Desktop installation](DESKTOP_INSTALLATION.md) documents download,
  first-open behavior, manual updates, uninstallation, data preservation, and
  known limitations.
- [CLI guide](CLI.md) covers installation, generation, profiles, automation,
  updates, and uninstallation.
- [Stack operations](STACK_OPERATIONS.md) covers starting, connecting,
  updating, backing up, and securing a CLI-generated media stack.
- [Development](DEVELOPMENT.md) covers local builds, validation, opt-in
  container contracts, CI, and release references.
- [Release procedure](RELEASING.md) documents unified versioning, the dry run,
  SBOM checks, and the protected publication approval.

## Accepted product direction

- [RFC 0001: Corsarr Desktop](rfcs/0001-corsarr-desktop.md) governs the product
  scope, user experience, runtime boundary, provisioning, security, operations,
  phased delivery, and validation plan.
- [ADR 0001: Wails v2](decisions/0001-use-wails-v2-for-desktop.md) records the
  accepted desktop framework.
- [ADR 0002: Runtime strategy](decisions/0002-docker-first-runtime-with-podman-gate.md)
  records the Docker-first MVP and the evidence gate for promoting Podman.
- [ADR 0003: Versioned quality profiles](decisions/0003-manage-desktop-quality-profiles-with-recyclarr.md)
  records the Desktop-only Recyclarr preset boundary and the separate advanced
  Profilarr direction.

The Desktop is a Wails v2 application in the existing Go module. It persists a
reviewed storage folder and application selection, validates the first Mac,
obtains versioned runtime consent, prepares or starts Docker Desktop on macOS,
and installs only digest-pinned containers owned by Corsarr. First launch is a
persisted sequential journey through welcome, authorization, environment,
storage, and application selection; it resumes after interruption and becomes
the regular lifecycle dashboard only after a complete installation. Its
lifecycle UI opens, starts, stops, restarts, updates, and removes each
application while keeping container removal separate from recoverable
configuration archival.
The reviewed movie/TV preset explicitly selects its supporting integrations.
Choosing an individual consumer marks its recommended integrations, but the
user can remove any suggestion after seeing which automatic setup will become
manual; installation still uses only the final reviewed selection. Installed
apps cannot be removed before their installed consumers or silently dropped
from desired state. Storage requires a fresh writable, measurable 10 GiB free
capacity check before preparation and installation, with missing hardlinks
reported as a non-blocking efficiency warning.

Provisioning currently covers Arr root folders and API keys, qBittorrent's
generated Keychain credential, paths and categories, Arr download clients,
Prowlarr connections, LazyLibrarian's generated administrator/API credentials,
book library, qBittorrent connection, and Prowlarr indexer synchronization,
Bazarr's Radarr/Sonarr connections, Jellyfin's local administrator and
libraries, and Seerr's Jellyfin/Radarr/Sonarr setup. FileFlows remains visible
for lifecycle management of an existing installation but cannot be selected
for one-click Desktop installation until it has an automated provisioning
adapter. Updates create a private configuration backup, replace the image,
verify readiness, and restore the previous container image on failure without
claiming to reverse a database migration.

When Radarr or Sonarr is selected, Desktop also presents a conditional quality
step. Four versioned Corsarr presets are synchronized by an ephemeral,
digest-pinned Recyclarr using a commit-pinned TRaSH Guides source; API keys are
injected only into the child process environment. An unmanaged choice is
available, and recurring synchronization remains off by default. The CLI does
not perform this post-install setup.

The Desktop also supports explicit Jellyfin LAN access and local-address
discovery, optional macOS login recovery for existing resources, redacted
diagnostic export, localized catalogs, and catalog-generated legal credits.
Windows/Linux onboarding and native secure stores on those platforms remain
open gates. The macOS release is intentionally ad-hoc signed and not notarized;
the installation guide records the resulting first-open requirement. The
empirical Podman promotion matrix also remains open.

The local macOS arm64 bundle builds, self-signs, declares macOS 14+, and has
passed deterministic suites plus visual smoke checks on the first Mac. No media
stack was installed during automated development validation: runtime terms,
image downloads, and application installation remain an explicit user action.

The Desktop also includes a catalog-generated "Aplicativos e licenças" screen.
It credits every offered application and image maintainer, identifies Docker and
Podman separately, shows the installed or approved digest, and opens only
backend-allowlisted official, source, license, image, and support links.
