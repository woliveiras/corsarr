# RFC: Corsarr Desktop for non-technical local users

- Status: accepted
- Authors: Codex and @woliveiras
- Created: 2026-08-10
- Decision owner: @woliveiras
- Target decision date: accepted on 2026-08-10

## Summary

Evolve Corsarr from a Docker Compose configuration generator into a desktop
application that installs and operates a focused media automation stack on a
user's personal computer. The user selects a storage location and applications;
Corsarr prepares the runtime, creates consistent volumes, installs and connects
the applications where supported, and provides lifecycle actions and links to
their original web interfaces.

The first MVP uses Docker behind a runtime abstraction. Podman is a candidate
for the default runtime, subject to the cross-platform evidence gate in
[ADR 0002](../decisions/0002-docker-first-runtime-with-podman-gate.md). The
desktop application uses Wails v2 as recorded in
[ADR 0001](../decisions/0001-use-wails-v2-for-desktop.md).

This RFC is accepted direction, not shipped behavior.

## Motivation and evidence

The current Corsarr CLI removes much of the Compose authoring burden, but a
non-technical user must still install and start Docker, run commands, understand
containers and virtual machines, and configure several independent web
applications. The current README also requires Docker and Compose as explicit
prerequisites.

Official sources establish that the remaining work is automatable within clear
boundaries:

- Docker exposes an Engine API and Go SDK for image and container lifecycle.
- Docker Desktop supports command-line installation and lifecycle operations.
- Microsoft WSL supports installation without adding a general-purpose Linux
  distribution, although elevation and restart may be required.
- Podman exposes native container, network, volume, secret, and machine
  operations without requiring Compose.
- The media applications expose web APIs or bootstrap configuration sufficient
  for substantial first-run automation, but credentials for external providers
  still require the user.

## Product principles

1. **Hide infrastructure concepts, not consent.** Primary UI copy says
   "Preparing the environment" and "Installing required components". Runtime
   names appear in consent, licenses, diagnostics, and technical details.
2. **Ask only for user-owned decisions.** The user chooses storage, selected
   applications, an administrator credential, and third-party provider
   credentials when needed.
3. **Keep original applications intact.** Corsarr automates installation and
   connection but opens the original web UI for normal and advanced use.
4. **Preserve data by default.** Removing an application never implies deleting
   its configuration, downloads, or library.
5. **Own desired state.** Corsarr owns the application catalog and reconciles it
   through a runtime adapter. Compose YAML is an implementation/export format,
   not the domain model.
6. **Do not bypass platform security.** UAC, macOS authorization, license
   acceptance, restart, firewall, and firmware requirements remain explicit
   human-authorized steps.

## Goals

- Provide a Windows, macOS, and Linux desktop application for a non-technical
  user running services on the same personal computer.
- Detect and prepare a supported container runtime through guided onboarding.
- Let the user choose one base storage location and create a safe shared layout.
- Install supported services from a versioned, approved catalog.
- Start, stop, restart, update, remove, and open each installed service.
- Apply one administrator username and password to compatible services where
  their supported bootstrap path allows it.
- Configure root folders, download-client categories, and inter-service API
  connections where supported and verified.
- Show plain-language progress and errors with expandable technical details.
- Provide complete application, image-maintainer, runtime, and third-party
  license credits.
- Keep the existing CLI and make it share the same Go application core.

## Non-goals

- Native installation of Radarr, Sonarr, Lidarr, Prowlarr, Bazarr, qBittorrent,
  Jellyfin, or related applications.
- Building a Linux distribution, custom operating system, or custom container
  virtual machine.
- Replacing the applications' complete web interfaces.
- Building a unified search, queue, history, playback, or media-library UI in
  the MVP.
- Single sign-on. Reusing an administrator credential is convenience, not a
  shared identity provider.
- Managing a NAS, mini-PC, remote Linux host, or cloud server in the MVP.
- Automatically configuring trackers, indexers, subtitle providers, VPN
  accounts, CAPTCHAs, two-factor authentication, or third-party terms.
- Preconfiguring sources that may facilitate copyright infringement.
- Supporting or replacing the archived Readarr project in the MVP.
- Automatically deleting media or downloads during application removal.

## User experience

### First run

1. Welcome and product/legal explanation.
2. System compatibility check.
3. Explicit consent to install and use the selected runtime and its supporting
   platform components.
4. Guided environment preparation, including elevation or restart when the OS
   requires it.
5. Resume at the same step after restart.
6. Choose one base storage folder.
7. Choose applications from the approved catalog.
8. Optionally choose one administrator username/password to apply to compatible
   applications.
9. Review storage, network exposure, versions, and licenses.
10. Install, provision, connect, and verify.

The main UI does not teach Docker, WSL, Podman, Compose, or virtual-machine
concepts. An error offers a plain-language action first and expandable technical
details second.

### Installed application card

Each installed application exposes only:

- Current state and a concise health/error indication.
- Open in browser.
- Start or stop.
- Restart.
- Update when an approved version exists.
- Remove application.
- Delete application data as a separate destructive action.

Closing Corsarr Desktop does not stop the media services. A separate user action
controls service shutdown.

### Removal semantics

`Remove application` stops and removes runtime resources owned by that
application while preserving its configuration and all shared media.

`Delete application data`:

- Identifies the exact Corsarr-owned configuration directory.
- Shows an approximate size.
- Offers a backup first.
- Requires an additional confirmation.
- Never includes shared downloads or libraries.

Deleting downloads or libraries is outside ordinary application removal and
requires a separate, explicit data-management flow if implemented later.

## Proposed architecture

### Repository and processes

Corsarr remains one product repository. Go owns the application core and the
Wails v2 backend; the desktop frontend uses TypeScript/HTML/CSS. Keep one Go
module until an independently versioned boundary is demonstrated.

```text
corsarr/
├── cmd/
│   ├── corsarr/                 # CLI entrypoint
│   └── corsarr-desktop/         # Wails entrypoint
├── internal/
│   ├── catalog/                 # approved app/runtime manifests
│   ├── orchestrator/            # install/update/remove workflows
│   ├── runtime/                 # runtime interface and adapters
│   │   ├── docker/
│   │   └── podman/
│   ├── services/                # per-application bootstrap adapters
│   ├── storage/                 # layout, capacity, permissions, backups
│   └── secrets/                 # OS credential-store boundary
├── frontend/                    # Wails web frontend
├── catalog/                     # versioned declarative catalog data
├── legal/                       # notices and license material
└── docs/
```

The existing directory layout may migrate incrementally; this tree is a target
boundary, not an instruction to move files before consumers exist.

### Desired-state model

The catalog describes applications independently of Docker Compose:

```go
type AppManifest struct {
    ID           string
    Image        ImageReference
    Ports        []PortBinding
    Mounts       []Mount
    Environment  map[string]string
    Dependencies []string
    HealthCheck  HealthCheck
    License      LicenseMetadata
}
```

The exact Go types remain implementation details, but the domain must represent:

- Versioned image references and immutable digests.
- Host/container mounts.
- Internal networking and user-facing port exposure.
- Dependencies and readiness.
- Runtime environment and secrets references.
- Application and image-maintainer attribution.
- Provisioning adapter/version compatibility.

The runtime boundary provides narrow lifecycle operations such as install,
start, stop, restart, inspect, logs, update, and remove. It does not accept an
arbitrary command from the frontend.

### Runtime topology

Use independent containers on one Corsarr-owned bridge network. Containers use
network aliases for service-to-service calls. Do not place the entire stack in a
single Podman pod because shared lifecycle and port publication would prevent
independent application updates and removal.

Administrative interfaces bind to loopback unless a service explicitly needs
LAN clients. Jellyfin LAN access is a separate, visible choice; opening every
administrative port to the LAN is not an acceptable default.

### Storage layout

The default layout is one host filesystem tree:

```text
Corsarr/
├── config/
│   ├── radarr/
│   ├── sonarr/
│   ├── lidarr/
│   ├── prowlarr/
│   ├── bazarr/
│   ├── qbittorrent/
│   └── jellyfin/
└── media/
    ├── downloads/
    │   ├── incomplete/
    │   └── complete/
    └── library/
        ├── movies/
        ├── tv/
        ├── music/
        └── books/
```

All containers see consistent internal paths under `/data`. Downloads and media
libraries must remain on the same filesystem/mount presented to the runtime so
hardlinks and atomic moves can work. Before installation, Corsarr tests capacity,
read/write permissions, and hardlink behavior. It explains a safe fallback when
the host filesystem cannot provide hardlinks.

Bind-mounted directories are preferred over opaque runtime volumes for user
data because they make ownership, backup, removal scope, and recovery explicit.

### Secrets and administrator credential

The optional shared administrator credential is applied separately to each
compatible application. It is not SSO.

- Store reusable credentials in the platform credential store.
- Never log passwords, API keys, VPN keys, or cookies.
- Do not put credentials in browser URLs.
- Prefer service APIs or documented bootstrap files over direct database edits.
- Keep secrets out of catalog manifests and non-encrypted profiles.
- Deliver a secret to only the adapter or service that needs it.

If an application cannot be bootstrapped safely, install it, report that one
configuration step remains, and open its original web UI.

## Provisioning workflow

Provisioning is idempotent and adapter-versioned. The target sequence is:

1. Create directories and secret references.
2. Create the Corsarr network.
3. Pull approved image digests.
4. Start Prowlarr, Radarr, Sonarr, and Lidarr as selected.
5. Wait for readiness and obtain supported API credentials.
6. Configure root folders.
7. Start qBittorrent and establish its administrator credential.
8. Create separate download categories for the selected managers.
9. Register qBittorrent with Radarr, Sonarr, and Lidarr.
10. Register selected applications with Prowlarr.
11. Start Bazarr and connect it to Radarr/Sonarr.
12. Start Jellyfin, create its administrator, and add selected libraries.
13. Run connection and path tests.

An adapter must distinguish `not yet supported`, `user input required`,
`temporarily unavailable`, and `failed`. Partial provisioning must not be called
complete. The user can still open an installed application's original UI.

## Windows onboarding

The Windows flow is an orchestrated installer, not a security bypass:

1. Check supported Windows version, architecture, memory, and virtualization.
2. Inspect WSL/runtime state without mutation.
3. Persist a resumable onboarding state before requesting elevation.
4. Show the component terms and obtain explicit consent.
5. If needed, request UAC elevation and run the official WSL installation with
   `--no-distribution`; use the documented web-download path when appropriate.
6. Detect a required restart and resume after the user returns.
7. Download the runtime installer from its official release endpoint.
8. Verify the published checksum/signature when available.
9. Run the supported unattended installer flags.
10. Start the runtime and wait for a healthy API before installing applications.

UAC confirmation, an OS-requested restart, and enabling virtualization in
BIOS/UEFI remain user actions. Corporate policy failures must produce a clear
unsupported/administrator-required result rather than a workaround.

Primary copy says "Preparing the environment". Docker, WSL, Podman, and the
specific commands remain visible in consent, licenses, diagnostics, and
expandable technical details.

## Compatibility and migration

### Delivery phases

#### Phase 1: reusable core and Wails shell

- Extract reusable application services without breaking the CLI.
- Add Wails v2 desktop entrypoint and frontend.
- Render the installed/available application catalog.
- Open an application URL through a safe Go method.

#### Phase 2: Docker lifecycle MVP

- Detect an existing supported Docker runtime.
- Install, start, stop, restart, remove, and open applications.
- Preserve data on removal.
- Implement the base storage layout and health/status reconciliation.

#### Phase 3: service provisioning and safe updates

- Add per-service bootstrap adapters.
- Connect applications and validate paths.
- Add approved version/digest updates, backup, verification, and rollback.
- Add licenses/credits UI generated from catalog metadata.

#### Phase 4: runtime onboarding

- Add guided Windows and macOS Docker Desktop installation with explicit terms.
- Add supported Linux runtime installation paths.
- Resume safely after elevation and restart.

#### Phase 5: Podman gate

- Execute the cross-platform spike from ADR 0002.
- Decide whether Podman becomes default, remains optional, or is rejected for a
  documented compatibility reason.

### Existing CLI

The CLI remains supported during the desktop migration. Its Compose generator
can initially feed the Docker runtime adapter. New domain behavior should move
behind shared Go packages rather than being reimplemented in the frontend.

Existing generated stacks are not automatically adopted in the MVP. Importing
or adopting unmanaged Compose projects requires a later design because Corsarr
must not claim ownership of resources it did not create.

### Rollback and cleanup

- A failed desktop milestone can leave the CLI unchanged.
- A failed application install removes only resources labeled as owned by that
  incomplete installation and preserves created user directories for diagnosis
  unless the user asks to delete them.
- Runtime uninstallation is never implied by removing the last application.
- Podman Machine or Docker Desktop data is never purged as ordinary cleanup.

## Security, privacy, and authority

### Trust boundaries

- Only the Go backend can access Docker/Podman sockets, named pipes, APIs, or
  local installer helpers.
- The Wails frontend receives a small allowlisted method surface.
- Managed containers never receive the runtime socket.
- Do not expose a runtime API over unauthenticated TCP, including localhost.
- Runtime resources carry Corsarr ownership labels; destructive operations
  resolve and verify those labels before acting.
- Application administration ports bind to loopback by default.

Docker and Podman both warn that their control API can provide arbitrary code
execution with the authority of the runtime user. Treat access as an
administrator-equivalent capability.

### Human-authorized operations

Corsarr may automate an operation only after the user has selected its scope.
The following remain explicit:

- Accept runtime and application terms.
- Approve UAC/macOS elevation.
- Restart the computer.
- Choose storage and LAN exposure.
- Provide third-party credentials.
- Start an application update.
- Delete application data.
- Delete downloads or media, if ever implemented.
- Uninstall or purge a runtime virtual machine.

### Legal and attribution

The application includes an "Applications and licenses" screen. For every
managed component it shows:

- Name, purpose, official site, and source repository.
- Application license and copyright notice.
- Container-image maintainer and image source.
- Installed version and image digest.
- Full license text and project support/donation link when official.
- A statement that Corsarr is not affiliated with the component project unless
  that relationship is explicitly established.

Generate `THIRD_PARTY_NOTICES` and a software bill of materials for distributed
desktop artifacts. Download runtime installers only from official sources,
verify available integrity metadata, and execute their official installer.
Bundling a runtime or offline application images requires a separate license,
trademark, dependency, and source-offer review.

Corsarr does not ship indexer credentials or preconfigure potentially unlawful
content sources. It may offer public-domain content for an end-to-end test.

## Reliability and operations

### Desired-state reconciliation

On launch and after runtime recovery, Corsarr compares its catalog/install state
with labeled runtime resources. Reconciliation is idempotent and does not adopt
or delete unrelated containers, networks, images, or directories.

On Windows/macOS with a runtime VM, Corsarr starts the machine/runtime first,
waits for its API, and then reconciles applications. Closing the UI does not
stop services. "Start services when I sign in" is an explicit setting.

### Updates

Do not track mutable `latest` tags as the installed contract. The catalog owns
approved versions and immutable digests.

Update flow:

1. Check capacity and runtime health.
2. Back up application configuration.
3. Pull the approved image.
4. Recreate the container with the same mounts/network/configuration.
5. Wait for health and run a minimal API check.
6. Retain enough prior runtime metadata to restore the previous image.
7. Report that an application database migration may prevent data rollback.

Automatic updates are opt-in. Runtime updates and application updates are
separate operations. Podman auto-update is not used in the MVP because it would
bypass the Corsarr catalog and verification flow.

### Diagnostics

User-visible errors include a plain-language state, next action, and expandable
technical detail. An exportable diagnostic bundle must redact secrets and may
include:

- Corsarr, catalog, runtime, OS, and architecture versions.
- Application/container state and health.
- Sanitized recent logs.
- Storage capacity, mount mapping, and permission test results.
- Failed workflow step and stable error code.

## Validation plan

### Deterministic tests

- Desired-state and manifest validation.
- Runtime adapter contract tests with fake and real local runtimes.
- Ownership-label checks before removal.
- Path normalization and traversal protection.
- Secret redaction.
- Update state machine, failed health check, and rollback behavior.
- Resume after simulated elevation/restart.
- Local-only administration port defaults.
- License/credits completeness for every catalog entry.

### Platform acceptance

Before a public MVP, validate the declared support matrix with clean machines or
VMs:

- Runtime absent, partially installed, stopped, outdated, and healthy.
- Virtualization absent or disabled.
- Restart required during onboarding.
- Default and external storage locations.
- Insufficient disk space and permission denial.
- Install, open, restart, update, remove-preserving-data, reinstall, and explicit
  data deletion.
- Host reboot/login recovery.
- Network offline during image pull or update.
- Corrupt/unhealthy application after update.

### Provisioning acceptance

For each supported application version, record which steps are fully automated,
which require user input, and which fall back to the original UI. A stack is
"configured" only when connection and path tests pass; container health alone
is insufficient.

## Alternatives

### Native application installation

- Benefits: No container runtime.
- Costs: Per-OS installers, services, paths, permissions, upgrades, rollbacks,
  and removers for every application.
- Why not selected: Explicitly out of product scope; it transfers an unbounded
  cross-platform maintenance matrix into Corsarr.

### Full unified media UI

- Benefits: One visual surface for search, queues, history, and libraries.
- Costs: Reimplements large and changing portions of every application API/UI.
- Why not selected: The MVP needs lifecycle and provisioning only; original web
  interfaces remain the correct owners of application behavior.

### Podman immediately

- Benefits: Direct native management, rootless operation, Apache-2.0 runtime,
  and no Compose requirement.
- Costs: Cross-platform machine paths, permissions, restart, and update behavior
  are not yet proven with Corsarr.
- Why not selected yet: ADR 0002 requires a bounded evidence spike before
  promotion.

### Custom VM and containerd distribution

- Benefits: Complete UX control.
- Costs: Corsarr becomes responsible for guest OS maintenance, virtualization,
  mounts, networking, updates, security response, recovery, and distribution.
- Why not selected: It would effectively create a container desktop product and
  contradict the decision not to create an operating system.

## Drawbacks and unresolved questions

- Final supported OS versions and CPU architectures for the first public MVP.
- Whether the initial vanilla TypeScript/CSS frontend needs a component library
  after the onboarding and lifecycle screens establish reusable UI patterns.
- Which service bootstrap endpoints/config files are stable enough to support
  per application version.
- Whether Docker Desktop may be redistributed; initial policy is download and
  execute the official installer after explicit consent.
- How macOS sandbox/file-access prompts interact with external media disks.
- Whether Windows NTFS/WSL and macOS shared mounts preserve acceptable hardlink
  behavior for the chosen runtime.
- Whether hardware transcoding belongs in the first Jellyfin catalog profile.
- Whether LAN access is limited to Jellyfin in the MVP.
- How runtime background startup is packaged on each OS without making the
  desktop window open unnecessarily.
- Which Podman result promotes it to default versus optional support.

## Official source catalog

These are research inputs, not blanket compatibility guarantees. Version-sensitive
claims must be rechecked during implementation and release validation.

### Desktop framework

- [Wails repository and v2/v3 status](https://github.com/wailsapp/wails)
- [Wails v2 supported platforms and build dependencies](https://wails.io/docs/v2.12.0/gettingstarted/installation/)
- [Wails Go/frontend binding model](https://wails.io/docs/howdoesitwork/)

### Docker

- [Docker Desktop installation on Windows and installer flags](https://docs.docker.com/desktop/setup/install/windows-install/)
- [Docker Desktop installation on macOS and installer flags](https://docs.docker.com/desktop/setup/install/mac-install/)
- [Docker Desktop CLI lifecycle](https://docs.docker.com/desktop/features/desktop-cli/)
- [Docker Engine API and Go SDK](https://docs.docker.com/reference/api/engine/sdk/)
- [Docker daemon socket protection](https://docs.docker.com/engine/security/protect-access/)
- [Docker rootless mode](https://docs.docker.com/engine/security/rootless/)
- [Docker Compose image pull behavior](https://docs.docker.com/reference/cli/docker/compose/pull/)
- [Docker Desktop terms summary](https://docs.docker.com/desktop/setup/install/windows-install/#docker-desktop-terms)
- [Docker legal terms](https://www.docker.com/legal/docker-terms-use/)

### Windows virtualization

- [Microsoft WSL installation](https://learn.microsoft.com/windows/wsl/install)
- [Microsoft WSL commands, including `--no-distribution` and `--web-download`](https://learn.microsoft.com/windows/wsl/basic-commands)

### Podman

- [Podman installation](https://podman.io/docs/installation)
- [Podman overview and remote API support](https://docs.podman.io/en/stable/)
- [Podman Machine lifecycle and platform providers](https://docs.podman.io/en/latest/markdown/podman-machine.1.html)
- [Podman Machine initialization, resources, and mounts](https://docs.podman.io/en/stable/markdown/podman-machine-init.1.html)
- [Podman REST API](https://docs.podman.io/en/stable/_static/api.html)
- [Podman system service behavior and security warning](https://docs.podman.io/en/stable/markdown/podman-system-service.1.html)
- [Podman Go bindings](https://pkg.go.dev/github.com/containers/podman/v5/pkg/bindings)
- [Podman container create options](https://docs.podman.io/en/latest/markdown/podman-create.1.html)
- [Podman network management](https://docs.podman.io/en/stable/markdown/podman-network-connect.1.html)
- [Podman volume management](https://docs.podman.io/en/latest/markdown/podman-volume.1.html)
- [Podman auto-update](https://docs.podman.io/en/stable/markdown/podman-auto-update.1.html)
- [Podman license](https://github.com/containers/podman/blob/main/LICENSE)

### Managed applications and images

- [Prowlarr application connections and synchronization](https://wiki.servarr.com/en/prowlarr/quick-start-guide)
- [Radarr authentication and API key](https://wiki.servarr.com/radarr/settings#security)
- [Radarr generated API-key configuration source](https://github.com/Radarr/Radarr/blob/develop/src/NzbDrone.Core/Configuration/ConfigFileProvider.cs)
- [Sonarr generated API-key configuration source](https://github.com/Sonarr/Sonarr/blob/develop/src/NzbDrone.Core/Configuration/ConfigFileProvider.cs)
- [Prowlarr generated API-key configuration source](https://github.com/Prowlarr/Prowlarr/blob/develop/src/NzbDrone.Core/Configuration/ConfigFileProvider.cs)
- [Lidarr generated API-key configuration source](https://github.com/Lidarr/Lidarr/blob/develop/src/NzbDrone.Core/Configuration/ConfigFileProvider.cs)
- [Radarr root-folder API controller and v3 contract](https://github.com/Radarr/Radarr/blob/develop/src/Radarr.Api.V3/RootFolders/RootFolderController.cs)
- [Sonarr root-folder API controller and v3 contract](https://github.com/Sonarr/Sonarr/blob/develop/src/Sonarr.Api.V3/RootFolders/RootFolderController.cs)
- [Lidarr root-folder API controller and v1 contract](https://github.com/Lidarr/Lidarr/blob/develop/src/Lidarr.Api.V1/RootFolders/RootFolderController.cs)
- [qBittorrent WebUI API](https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-%28qBittorrent-5.0%29)
- [qBittorrent API key authentication](https://github.com/qbittorrent/qBittorrent/wiki/API-Key-Authentication-%28%E2%89%A5v5.2.0%29)
- [Jellyfin server API and local Swagger UI](https://github.com/jellyfin/jellyfin)
- [Jellyfin setup wizard and administrator](https://jellyfin.org/docs/general/post-install/setup-wizard/)
- [Jellyfin official container image](https://jellyfin.org/docs/general/installation/container/)
- [Bazarr supported container images](https://wiki.bazarr.media/Getting-Started/Installation/Docker/docker/)
- [LinuxServer Radarr image, volumes, permissions, and hardlink warning](https://docs.linuxserver.io/images/docker-radarr/)
- [Sonarr root folders, downloads, and hardlinks](https://wiki.servarr.com/en/sonarr/quick-start-guide)
- [Archived Readarr repository](https://github.com/Readarr/Readarr)

### Application licenses

- [Radarr GPL-3.0](https://github.com/Radarr/Radarr/blob/develop/LICENSE)
- [Prowlarr GPL-3.0](https://github.com/Prowlarr/Prowlarr/blob/develop/LICENSE)
- [Lidarr GPL-3.0](https://github.com/Lidarr/Lidarr/blob/develop/LICENSE.md)
- [Bazarr GPL-3.0](https://github.com/morpheus65535/bazarr/blob/master/LICENSE)
- [qBittorrent GPL licenses](https://github.com/qbittorrent/qBittorrent)

## Decision and implementation history

- 2026-08-10: Accepted the focused local-desktop product scope.
- 2026-08-10: Rejected native per-application installation.
- 2026-08-10: Accepted Wails v2 for the desktop application.
- 2026-08-10: Accepted Docker-first delivery behind a runtime boundary.
- 2026-08-10: Defined the cross-platform Podman promotion gate.
- 2026-08-10: Deferred Readarr/replacements and remote-host management.
- 2026-08-10: Started Phase 1 with a reusable application catalog and Wails v2
  shell; built and visually smoke-tested the first self-signed macOS arm64 app.
- 2026-08-10: Started Phase 2 with bounded read-only Docker diagnostics and a
  native storage picker that checks writing, hardlinks, and available space.
- 2026-08-10: Added persistent reviewed setup, automatic catalog dependencies,
  and explicit idempotent creation of the shared storage layout. Container
  installation and service provisioning remain pending.
- 2026-08-10: Added the validated runtime contract, ownership-safe Docker
  lifecycle adapter, and the first multi-architecture image-digest catalog.
  These boundaries are not exposed as an install action yet.
- 2026-08-10: Added versioned runtime consent to persistent desktop state.
  Existing setup state migrates without being treated as accepted consent.
- 2026-08-10: Connected the reviewed setup to the transactional Docker install
  flow. This installs containers only with an existing healthy runtime;
  application provisioning and runtime onboarding remain pending.
- 2026-08-10: Exposed catalog-scoped application status and start, stop,
  restart, and remove-preserving-data actions in Corsarr Desktop.
- 2026-08-10: Separated container removal from recoverable application-data
  removal; shared media and downloads are never deletion targets.
- 2026-08-10: Added bounded, redirect-free readiness checks against catalog
  loopback endpoints before installation advances to dependent applications.
- 2026-08-10: Added bounded, symlink-safe local discovery of generated Arr API
  keys with redacted formatting and no frontend exposure.
- 2026-08-10: Added idempotent authenticated creation of approved Radarr,
  Sonarr, and Lidarr library root folders after readiness.
- 2026-08-10: Added ownership-checked, bounded backend log access for secret
  bootstrap without exposing a general log method to the frontend.
