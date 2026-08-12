# Corsarr CLI Architecture

> 🏴‍☠️ Navigate the high seas of media automation

> This document primarily describes the current CLI and Docker Compose generator.
> The accepted Corsarr Desktop direction is documented in
> [RFC 0001](rfcs/0001-corsarr-desktop.md). Durable desktop technology and
> runtime choices are recorded under [decisions](decisions/). Its first
> development slice is described below and is not part of a released version.

## Corsarr Desktop development slice

The first native shell lives under `desktop/` and is a second `main` package in
the existing Go module. Wails v2 embeds the production frontend and exposes a
small Go-bound method surface. The frontend can list catalog applications and
request that an application be opened by ID. It can request predefined,
read-only environment diagnostics and bounded catalog application intents, but
cannot submit a URL, shell command, runtime argument, image, mount, or container
name.

The Wails surface now also exposes explicit current-terms acceptance and one
bounded `InstallSelectedApplications` intent. That method reads only persisted,
reviewed setup and delegates to `InstallationService`; the frontend cannot
provide an image, mount, container name, runtime argument, or cleanup target.
An application-wide non-blocking change gate wraps every setup, runtime, storage
layout, update, lifecycle, and configuration-archive intent. A second mutation
is rejected before it reaches an adapter, and login recovery uses the same gate;
read-only status and catalog methods remain available while work is in progress.
Result DTOs also minimize backend disclosure: installation returns a catalog ID
plus a bounded failure flag, update returns updated/rolled-back/attention flags,
and configuration archival returns only whether it occurred. Recovery paths,
backup checksums, runtime status objects, and raw backend errors are retained for
Go control flow but carry `json:"-"` and do not appear in generated bindings.

Desktop-state schema 7 owns the one-time setup journey independently of the
frontend. It records the canonical Desktop language and the furthest completed
onboarding step, and never backfills completion during migration. The
application service enforces the
sequence welcome, authorization, environment, storage, applications: current
terms gate environment preparation, fresh runtime and storage checks gate their
respective transitions, and only a complete successful installation marks the
journey complete. An interrupted launch therefore resumes safely, while local
Back navigation can explain an earlier screen without weakening those guards.

The Desktop localization boundary uses canonical locale identifiers `en`, `es`,
`pt-BR`, and `it`. On first launch the frontend maps the first supported system
locale, falls back to English, and persists the result through
`SetupService.SaveLanguagePreference`. Later launches treat the Go state as
authoritative; local storage is only a startup cache that avoids rendering in
the previous language before the Wails bridge is ready. The frontend bundles
i18next resources and checks exact key parity across all four catalogs. Go
continues to own the CLI and service-description catalogs, whose YAML keys are
also checked for parity. Backend operation contracts expose stable issue codes;
the frontend translates known codes and uses backend prose only as a fallback
for a code introduced by a newer backend.

`internal/application.Catalog` is the first presentation-independent application
service. It derives user-facing entries from `internal/services.Registry`,
excludes infrastructure-only services without web UI metadata, and resolves
only allowlisted loopback URLs. The CLI remains unchanged and continues to use
the existing generator directly while reusable application services are
extracted incrementally.

`internal/runtime` defines the runtime-neutral probe contract and the first
Docker detector. It resolves the executable through the host, runs only fixed
`docker --version` and `docker info` checks under a five-second deadline, and
distinguishes an unavailable client, a stopped runtime, a ready runtime, and an
unexpected error. Raw runtime access remains inside Go; the frontend receives a
bounded status object.

`internal/onboarding` implements the first macOS runtime preparation path. A
stopped Docker Desktop is started with the official Desktop CLI, with a fixed
`open` fallback for older installations. When absent, Corsarr downloads the
architecture-specific Docker Desktop 4.86.0 disk image directly from Docker's
versioned endpoint into a private temporary cache, enforces a 2 GiB limit,
verifies the official SHA-256, mounts it read-only, verifies the Apple code
signature, Docker Team ID `9BNSXJN65R`, signed contents, and Gatekeeper policy,
then invokes Docker's official installer with `--accept-license` and `--user`
through the native macOS administrator authorization prompt. The temporary disk
image and mount workspace are removed afterward. Runtime readiness is probed
for a bounded three minutes before success is reported.

`internal/autostart` is the explicit background-start boundary. On macOS 13+
it uses Apple's `SMAppService.mainAppService` to register the main Corsarr app
for subsequent logins, report when approval is still required, and open the
system Login Items panel. It does not write a legacy LaunchAgent plist, request
administrator rights, or register anything until the user enables the option.
Other platforms remain unsupported until they have native, separately tested
adapters.

The preference is stored in desktop-state schema 3 only after Service
Management accepts the change. UI state is reconciled from the native service
status, so a change made in System Settings is not falsely reported as active.
An approval-required registration is represented separately from an enabled
registration and offers a bounded action to open the correct System Settings
panel.

When that native setting is enabled, startup runs a non-installing recovery in
the background. `DockerService.Recover` can start an existing Docker Desktop
but has no installer path. `application.RecoveryService` then orders selected
related applications and starts only stopped Corsarr containers
that already exist. Missing containers are skipped; images are not pulled,
containers are not created, provisioning is not rerun, and no resource is
removed. A bounded Wails event asks the frontend to refresh status after the
attempt without sending technical errors or secrets.

`PrepareRuntime` is available only after the current versioned runtime consent
has been persisted. The frontend supplies no URL, checksum, command, path,
username, or installer argument. Windows and Linux return an explicit unsupported
result until their separately tested onboarding implementations ship.

The package also owns the validated `ContainerSpec` boundary used by future
runtime adapters. A spec contains resolved values rather than templates or
commands. It requires an immutable `sha256` image reference, a safe catalog ID,
absolute bind mounts, valid ports, and an explicit loopback-or-LAN exposure for
every published port. The Docker and Podman translation layers must consume this
contract; it is not exposed as a general-purpose Wails method.

`internal/runtime.DockerManager` is the first adapter for that contract. It uses
fixed Docker CLI operations with argument arrays, creates a labeled bridge
network, translates validated specs into labeled containers, and supports
inspect/start/stop/restart/remove. Lifecycle changes verify Corsarr ownership
before touching an existing container; a resource with the expected name but
without matching labels is rejected. Application services call this adapter
only after catalog, reviewed-setup, and consent checks.

`internal/runtime.PodmanManager` implements the same contract with direct,
fixed Podman CLI operations. It manages independent containers on the same
labeled network; it does not use Compose or place the media stack in a shared
pod. The adapter exists for the accepted cross-platform spike and is not wired
as the desktop default. Podman installation, Podman Machine supervision, host
path translation, and the empirical promotion matrix remain separate gates.

The spike also has a bounded read-only `PodmanDetector` and a separate
`PodmanMachineService`. On macOS and Windows, that service inspects only the
fixed rootless `podman-machine-default`, initializes it with Corsarr's reviewed
resource defaults when absent, starts it when stopped, and polls runtime
readiness. It never removes or purges a machine, changes its provider, installs
the Podman client, or invokes Podman Machine on native Linux. Neither component
is exposed through Wails while the promotion gate remains open.

The adapter can also return at most 500 trailing log lines from a container,
after the same ownership verification. This capability is backend-only and is
not exposed through Wails because startup logs may contain temporary secrets;
its first intended consumer is qBittorrent credential bootstrap.

`internal/catalog.RuntimeCatalog` is the approved desktop translation from the
existing service registry to `ContainerSpec`. Its image references are pinned
to multi-architecture OCI index digests verified on 2026-08-10, and each entry
keeps the image maintainer's installation source. It resolves one private
configuration mount per application plus the single shared media mount. Web
administration ports remain loopback-only in the default spec. Catalog refresh,
review, and update policy remain separate from ordinary application startup.

The user-facing `internal/application.Catalog` keeps its existing English
constructor for CLI compatibility. Corsarr Desktop uses
`NewLocalizedCatalog` with the embedded Brazilian Portuguese locale, so names
and descriptions come from the same reviewed translation keys already used by
the CLI instead of being duplicated in TypeScript. Missing translation keys
fall back to registry metadata.

The desktop legal catalog applies the same locale to application names and
purposes while retaining allowlisted official, source, image, support, and
license URLs in Go. Runtime legal notices remain authored explicitly because
they are not service-registry entries.

`RuntimeOptions.AllowJellyfinLAN` is the only MVP exception to the loopback
default. It changes Jellyfin's approved TCP 8096 binding to explicit LAN
exposure while every other application remains on loopback. The persisted
choice is accepted only while Jellyfin is selected and before its container
exists; changing it later requires removing only the container and reinstalling
with the preserved configuration/media. The frontend asks for confirmation
before enabling it.

`internal/localnetwork.Discoverer` enumerates active, non-loopback host
interfaces and returns only private IPv4 URLs for Jellyfin TCP 8096. Known
container, VM, tunnel, and macOS peer-to-peer interface prefixes are excluded;
failure to enumerate produces no address rather than a guessed public URL. The
Wails boundary exposes these URLs only while the persisted Jellyfin LAN choice
is enabled, and its clipboard intent accepts only an exact address rediscovered
by the Go backend.

`internal/hostprofile.Profiler` derives the shared runtime identity and
timezone once when the desktop backend starts. Native Linux uses the current
positive UID/GID so bind-mounted files remain owned by the desktop user.
macOS and Windows use the stable `1000:1000` identity inside their Linux VM
instead of leaking host identifiers that have different semantics. Timezone
detection accepts only loadable IANA names from `TZ`, the `/etc/localtime`
target, or the Go local location and otherwise falls back to `Etc/UTC`. These
defaults are applied to every approved image that supports `PUID`, `PGID`, and
`TZ`; the reviewed Jellyfin network choice is layered on without replacing
them.

`internal/orchestrator.Installer` owns the first transactional application
workflow: resolve and validate the approved spec, ensure the network, pull,
create, start, and inspect. A failure after container creation removes only that
owned incomplete container using a non-canceled cleanup context. Bind-mounted
configuration and media are deliberately outside cleanup. The orchestrator is
covered through the runtime interface and reached from Wails only through the
bounded `InstallSelectedApplications` intent.

The desktop install intent is end-to-end: its backend rechecks current terms,
prepares or starts the supported runtime, requires a ready result, and only then
delegates to `InstallationService`. When the runtime is absent, the frontend
also asks for a final explicit confirmation before the pinned installer is
downloaded. A preparation failure never reaches container installation.

`internal/hostreadiness` owns the read-only first-Mac preflight. It measures
the macOS version, physical memory, architecture, and free bytes on the
runtime-cache filesystem under a three-second deadline. The gate requires
macOS 14+, `arm64` or `amd64`, Docker's documented minimum 4 GiB RAM, and a
Corsarr operational margin of 4 GiB free (twice the installer's 2 GiB download
limit). Missing measurements fail closed. `EnvironmentService` exposes the
facts for explanation, while `PrepareRuntime` and the combined install intent
recheck the same backend gate immediately before mutation.

Installation is reconciliatory: a matching running container is reused, and a
matching stopped container is started, but only when its recorded runtime
contract fingerprint also matches the resolved approved spec. Missing or
divergent fingerprints require container removal with data preservation and a
fresh install; a matching image alone is insufficient. A differently pinned
image is never replaced implicitly by installation; that case is routed to the
explicit update/backup/rollback flow described below.
`internal/application.InstallationService` enforces current consent, prepares
the reviewed layout, orders selected integrations before consumers, and returns
a structured per-application result while preserving already completed apps.
Its optional observer emits only catalog ID, bounded stage, position, and total
for desktop progress. Runtime output and provisioning errors are excluded from
events so progress reporting cannot become a log or credential channel. The
onboarding frontend switches to a dedicated installation view before invoking
the backend. It keeps a local per-application view from those events, showing
unstarted selections as waiting and retaining ready or failed states for the
duration of the attempt. After the last application is ready, a separate final
phase remains active while onboarding integrations and any managed quality
profile are applied. An interrupted attempt keeps the technical issue available
and exposes retry without discarding the user's selections. Application,
quality-profile, and final reconciliation failures also produce a bounded,
redacted support log that crosses the desktop bridge only when the user chooses
to copy it; runtime environment secrets and user paths are removed first.
`internal/provisioning.HTTPReadiness` then probes only the catalog-resolved
loopback URL, without credentials or redirects, until the application accepts
HTTP or a bounded timeout expires. A newly created container that never becomes
ready is removed while its bind-mounted data remains available for diagnosis;
an existing container is never removed by this check.

`internal/orchestrator.Updater` owns container replacement. It first verifies
that the current owned container uses an immutable image that can be restored,
and that its recorded SHA-256 contract fingerprint matches the approved spec
after excluding only the image reference. Port, mount, init, application, or
environment drift therefore fails before backup, pull, stop, or removal instead
of pretending that an old image plus a new runtime contract is a rollback. It
creates the private configuration backup, pulls the approved digest, and only
then replaces the container. It starts the replacement for bounded readiness
verification and preserves whether the prior container was running or stopped.
Any create, start, inspect, or readiness failure removes the replacement and
recreates the previous image under a non-canceled cleanup context. A successful
container rollback preserves service availability and the backup artifact, but
does not claim to reverse an application database migration.

`internal/provisioning.ARRCredentialReader` reads the generated `ApiKey` only
from a fixed `config/<known-arr>/config.xml` beneath the reviewed Corsarr root.
It rejects unknown apps, symlinked paths, oversized XML, and malformed keys.
The returned credential redacts default formatting and JSON serialization and
is never part of the Wails API; only the Go provisioning layer may explicitly
reveal it for an allowlisted loopback API request.

`internal/provisioning.ARRClient` accepts only the catalog loopback endpoint,
the correct API version, and the approved container library path for Radarr,
Sonarr, or Lidarr. It sends the key in `X-Api-Key`, disables proxy and redirect
following, bounds response bodies, lists existing root folders first, and posts
only when the desired path is absent. `ARRProvisioner` connects this client to
installation after readiness, making retries idempotent. Unsupported apps are
left unchanged for their dedicated provisioners.

Before other Arr provisioning, `ARRAuthenticationProvisioner` inspects the live
host configuration and acts only on the exact pristine state: no authentication
method and no username. It creates a unique password for each supported Arr
application, stores it in the native credential store, enables Forms
authentication for `corsarr`, and keeps loopback access exempt so backend
automation can continue without browser sessions. Existing or concurrently
changed authentication is never overwritten. If the application accepted the
credential but final verification was interrupted, the credential remains in
the Keychain so the user is not locked out. Subsequent desktop startup
reconciles this setup for already-running managed containers without pulling,
creating, starting, or reinstalling them. Enabling start at login separately
authorizes the runtime and existing containers to be started first.

`internal/credentials.Store` is the boundary for generated service secrets.
The first platform adapter uses the macOS Keychain with a fixed service name
and allowlisted account keys. Secret values redact default formatting and JSON,
never enter desktop state, and are not exposed by Wails. Other platforms return
an explicit unsupported error until their native secure-store adapters ship.

`internal/provisioning.QBittorrentProvisioner` consumes the bounded owned log
tail only when no stored credential exists, extracts the official temporary
password without logging it, authenticates as `admin`, generates and stores a
permanent password before activating the `corsarr` user, then verifies a fresh
login. A failed activation removes the prepared Keychain entry. On every
reconcile it enforces the shared complete/incomplete paths and the approved
LazyLibrarian, Radarr, Sonarr, and Lidarr categories. The HTTP client is loopback-only,
redirect-free, cookie-scoped, and response-bounded.

The Wails surface can report only whether qBittorrent access is available and
the non-secret username. Password retrieval remains in Go: an explicit
`CopyQBittorrentPassword` intent writes it directly to the native clipboard and
returns no secret to TypeScript.

The same boundary applies to managed Arr credentials. The dashboard receives
only application ID, username, and availability. An explicit allowlisted
`CopyARRPassword` intent loads one app-specific secret in Go and writes it
directly to the native clipboard; the password and Arr API key never cross the
Wails return boundary.

The same native-clipboard boundary exposes the generated LazyLibrarian
administrator password. Its API key remains backend-only.

`internal/diagnostics.Reporter` builds the support snapshot only after the user
chooses Export diagnostics. It includes bounded platform, runtime, catalog,
application, setup, and storage facts, redacts credential-shaped text, and
deliberately excludes logs, cookies, request bodies, runtime sockets, passwords,
and API keys. `FileWriter` accepts only an absolute user-selected destination,
rejects symlinks/non-regular targets, writes mode `0600`, syncs, and atomically
renames the JSON. Canceling the native save dialog performs no collection or
write.

`internal/provisioning.ARRDownloadClientProvisioner` loads both secrets only in
Go and asks `ARRClient` to reconcile the reserved `qBittorrent (Corsarr)`
provider. The client first fetches the app's official live provider schema,
requires every expected qBittorrent field, fills only the internal network
alias and runtime-aligned WebUI port `8081`, stored credentials, and
app-specific category, then creates or updates only the exact reserved provider
name. Other user providers are never selected. Radarr and Sonarr use API v3;
Lidarr uses API v1.

`internal/provisioning.LazyLibrarianProvisioner` creates separate administrator
and API credentials in the native store, then submits only the reviewed
configuration fields supported by the pinned service to its allowlisted
loopback endpoint. It enables Basic
authentication, fixes the book library and download paths, connects the
managed qBittorrent user/category when selected, restarts the service so the
web authentication boundary takes effect, and verifies both the API and
qBittorrent connection. Retries reuse the stored credentials.

`internal/provisioning.ProwlarrProvisioner` runs when each supported target app
becomes ready and Prowlarr is also selected. For Arr targets it reads both keys
from their fixed config files. For LazyLibrarian it loads the generated API key
and administrator password from the native store. It reconciles only
`<App> (Corsarr)` through Prowlarr API v1. The client
starts from the live application schema, retains the default category lists for
Arr targets, sets reviewed book categories for LazyLibrarian, enables full sync
plus the internal `prowlarr` and target network URLs, and relies
on Prowlarr's provider create/update path to validate connectivity. User-created
Prowlarr applications remain untouched.

`internal/provisioning.BazarrProvisioner` connects Bazarr only when Radarr and
Sonarr are also selected and ready. It reads Bazarr's generated API key only
from the fixed `config/bazarr/config/config.yaml` path and reuses the redacted
Arr credential boundary for the other keys. `BazarrClient` submits only the
documented settings fields to the loopback-only `/api/system/settings`
endpoint, uses internal network aliases, then reads the settings back and
verifies that both connections were persisted. No API key crosses the Wails
boundary.

`internal/provisioning.JellyfinProvisioner` owns the first-run wizard. It stores
a generated `corsarr` administrator password in the platform credential store
before activation and removes a newly generated value only when Jellyfin never
accepted it. `JellyfinClient` detects the public wizard state, authenticates
through the official user endpoint, creates only the reserved movie, TV, and
music libraries under `/data/library`, disables remote access by default, and
then completes the wizard. Existing matching libraries are preserved; a user
library occupying a reserved Corsarr name with different settings is rejected.
The HTTP client is loopback-only, proxy-free, redirect-free, and response
bounded. The desktop can copy the stored password directly to the native
clipboard without returning it to TypeScript.

`internal/provisioning.SeerrProvisioner` configures Seerr after its selected
Jellyfin, Radarr, and Sonarr integrations are ready. It loads their credentials
only in Go, authenticates Seerr through the official Jellyfin login route, and retains the resulting HTTP-only
session in a private cookie jar. `SeerrClient` discovers/enables Jellyfin
libraries, tests both Arr connections to obtain live profiles and root folders,
prefers a `Corsarr - ` quality profile, then `Any` (or the lowest returned ID), and reconciles
only the reserved `<App> (Corsarr)` entries with safe defaults. It initializes
Seerr only after every connection succeeds. The client is loopback-only,
proxy-free, redirect-free, and response-bounded.
Because Seerr authenticates through Jellyfin, its Desktop card reuses the
Jellyfin username and explicit native-clipboard password action instead of
creating or exposing a separate credential.

`internal/quality.Syncer` owns the optional Desktop quality-profile boundary
described by [ADR 0003](decisions/0003-manage-desktop-quality-profiles-with-recyclarr.md).
After selected Arr applications are ready, it generates private Recyclarr files
from a versioned Corsarr preset, references API keys through environment
variables, and executes a digest-pinned ephemeral Recyclarr container first in
preview mode and then in apply mode. The TRaSH Guides resource provider is
commit-pinned. Raw command output and secrets do not enter the Wails response.
No scheduler is created, and the CLI remains outside this automation boundary.

`internal/application.ManagementService` translates runtime inspection into
`not_installed`, `running`, `stopped`, or `attention` for every catalog app. Its
start, stop, restart, and remove intents validate the catalog ID before reaching
the runtime. It also compares an installed image with the approved catalog
digest so the UI can offer an update only when one exists. Remove deletes only
the labeled container; the bind-mounted config and shared media tree are outside
its authority. Before removal it re-inspects direct catalog dependents and
rejects the operation while any consumer remains installed. The same blockers
are exposed as bounded IDs for an explanatory desktop state.

`internal/application.UpdateService` accepts only one catalog application ID,
reloads the persisted storage choice and current consent, serializes update
attempts, and supplies the runtime options internally. It invokes provisioning
again only after a verified replacement. Runtime/rollback failures are returned
as a structured result so the desktop can distinguish a restored previous image
from an update that needs attention. The Wails surface cannot provide an image,
backup path, mount, or runtime argument.

`internal/legal.Catalog` is a build-time completeness gate and the single source
for the desktop credits screen. Construction fails when a user-facing catalog
application lacks legal metadata or approved-image attribution. Each entry
contains the project purpose, license identification, copyright and
non-affiliation notices, image maintainer, approved digest, and semantic link
labels. URLs remain private to Go: `OpenLegalLink` accepts only a component ID
plus an allowlisted link kind and resolves an HTTPS URL internally. Docker
Desktop and the Podman candidate are listed explicitly alongside the managed
applications.

The dashboard's read-only Info screen obtains its installed Corsarr version
from `internal/buildinfo` and its quality-policy, Recyclarr, and pinned TRaSH
Guides versions from `internal/quality`. The Wails DTO contains only these
bounded identifiers plus whether automatic profile updates are enabled. This
keeps implementation metadata out of onboarding and prevents the frontend from
guessing the release version or duplicating synchronization constants.

`internal/application.DataManagementService` is a distinct destructive-action
boundary. It requires the catalog application container to be absent, reloads
the persisted storage location, and delegates only that application's config
directory to `internal/storage.ApplicationDataManager`. The storage manager
refuses unsafe IDs and symlink targets. Its read-only status walk sums only
regular files inside that application directory, rejects nested links and
special files, and exposes the approximate byte count for the final desktop
confirmation. It then atomically moves configuration into
the private `<selected>/Corsarr/trash/config/<app>/` tree. It never receives or
targets the shared media and downloads paths. qBittorrent and Jellyfin Keychain
entries are removed only after their configuration becomes recoverable. If
credential deletion fails, the storage adapter restores the archived directory
to its exact original location before reporting failure.

`internal/storage.BackupManager` is the update workflow's recovery boundary. It
accepts only a reviewed Corsarr root plus a safe catalog application ID and
archives exactly `config/<application>` into a private `tar.gz` below
`backups/config/<application>`. It refuses symlinks and special files, streams
the archive without loading it into memory, computes SHA-256 over the compressed
artifact, and publishes it atomically with mode `0600`. Media paths are not part
of this API. Creating this recovery point does not claim that an older container
can reverse an application database migration; the application surface must
communicate that limitation independently.

`internal/storage` inspects only a directory returned by the native Wails folder
picker. It verifies that the path already exists and is a directory, creates
temporary write and hardlink probes, reports available space, and removes all
probe artifacts. A ready selection is persisted through
`internal/application.SetupService` and `internal/state.FileStore`; rejected or
canceled selections are not persisted. The state file contains only the storage
path and approved application IDs, uses the operating system's user
configuration directory, and is written with private file permissions where the
platform supports them. A folder is ready only when free capacity is measurable
and at least 10 GiB; hardlink failure remains a visible efficiency warning.
Both folder preparation and installation repeat this inspection before mutation
so a disconnected external disk, permission change, or newly full filesystem
cannot pass on stale setup state.
The desktop update intent repeats the same inspection before entering
`UpdateService`, so no backup, image pull, stop, or container replacement begins
from stale storage approval.
It then repeats `hostreadiness.Check` to cover the runtime-cache filesystem on
the Mac, which may be different from an external selected media disk. A failed
host margin likewise stops before `UpdateService`.

The state schema also records the accepted runtime-terms version and UTC
timestamp. Schema 1 setup files migrate to schema 2 without inventing consent.
Folder preparation remains available after storage/application review, but the
backend reports installation authority only when the current terms version was
explicitly accepted. A future terms version therefore requires new consent.

Application selection is validated against the presentation-safe catalog and
the exact deterministic selection is persisted. The frontend uses those catalog
relationships as reversible recommendations: selecting a consumer recursively
marks its integrations, deselecting any item removes only that item, and an
inline advisory names the automation that will become manual. Catalog
relationships order compatible applications only when both remain selected;
provisioning likewise skips an absent integration so existing external services
remain the user's choice. The reviewed movie/TV preset explicitly selects its
complete integrated stack and is owned by the Go catalog, not duplicated in
TypeScript. Running and stopped applications remain selected;
during a runtime outage only previously persisted uncertain selections are
retained, avoiding both accidental deselection and the false assumption that
every catalog app is installed. The frontend cannot provide an arbitrary
directory name to the layout operation: it can only request preparation from
the persisted, reviewed setup.

`internal/storage.LayoutPreparer` validates all application IDs before writing
and idempotently creates `<selected>/Corsarr/config/<app>` plus one shared media
tree for downloads and libraries. Configuration directories are private; an
existing selected folder and unrelated files are preserved. This slice creates
directories only; installation, lifecycle, and recoverable configuration
removal live behind the separate application services described above.

## 📋 Overview

Corsarr currently ships as a Go CLI that simplifies configuration and
initialization of the *arr stack. Users can select services, configure shared
environment values, validate the result, and generate `docker-compose.yml` and
`.env` files.

The CLI replaced the original model of maintaining separate hand-edited Compose
files for VPN and non-VPN stacks. It currently:

1. Allows visual service selection (checkboxes)
2. Configures environment variables via prompts
3. Generates files automatically based on choices
4. Validates configurations before creating files
5. Supports profiles for configuration reuse

---

## 🏗️ Project Architecture

```
corsarr/
├── cmd/
│   ├── root.go           # Main command and Cobra setup
│   ├── generate.go       # Command to generate docker-compose and .env
│   ├── preview.go        # Preview configurations before generating
│   ├── profile.go        # Manage saved profiles (save/load/list)
│   ├── health.go         # Check container health status
│   └── check_ports.go    # Check port conflicts
│
├── internal/
│   ├── services/
│   │   ├── services.go   # Definition of all available services
│   │   ├── categories.go # Service categorization
│   │   ├── registry.go   # Registry pattern to manage services
│   │   └── templates/    # Embedded YAML service definitions
│   │
│   ├── generator/
│   │   ├── compose.go    # docker-compose.yml generation orchestrator
│   │   ├── strategy.go   # Strategy Pattern (VPN/Bridge mode)
│   │   ├── env.go        # .env file generation
│   │   ├── network.go    # Docker network configuration
│   │   └── templates/    # Embedded generation templates
│   │
│   ├── validator/
│   │   ├── validator.go  # Configuration validations
│   │   ├── port.go       # Port conflict validation
│   │   ├── dependency.go # Service dependency validation
│   │   ├── path.go       # Path validation
│   │   ├── path_unix.go  # Unix-specific disk space checking
│   │   ├── path_windows.go # Windows-specific disk space checking
│   │   └── docker.go     # Docker installation validation
│   │
│   ├── prompts/
│   │   ├── interactive.go # Interactive prompts (Huh/Bubble Tea)
│   │   └── config.go      # Environment variable config prompts
│   │
│   ├── profile/
│   │   └── profile.go     # Profile structure and persistence
│   │
│   └── i18n/
│       ├── i18n.go        # Internationalization
│       ├── language.go    # Language selection and normalization
│       └── locales/       # Translation files
│           ├── en.yaml    # English
│           ├── pt-br.yaml # Brazilian Portuguese
│           └── es.yaml    # Spanish
│
├── profiles/            # Directory for saved profiles
├── .goreleaser.yml      # GoReleaser configuration
├── go.mod
├── go.sum
├── main.go
└── README.md
```

---

## 🔧 Identified Services

### Download Managers
| Service | Port | Image | Present in |
|---------|------|-------|------------|
| qBittorrent | 8081 | lscr.io/linuxserver/qbittorrent:latest | VPN, Simple |

### Indexers
| Service | Port | Image | Present in |
|---------|------|-------|------------|
| Prowlarr | 9696 | lscr.io/linuxserver/prowlarr:latest | VPN, Simple |
| FlareSolverr | 8191 | ghcr.io/flaresolverr/flaresolverr:latest | VPN |

### Media Management
| Service | Port | Image | Present in |
|---------|------|-------|------------|
| Sonarr (TV) | 8989 | lscr.io/linuxserver/sonarr:latest | VPN, Simple |
| Radarr (Movies) | 7878 | lscr.io/linuxserver/radarr:latest | VPN, Simple |
| Lidarr (Music) | 8686 | ghcr.io/hotio/lidarr:latest | Simple |
| LazyLibrarian (Books) | 5299 | lscr.io/linuxserver/lazylibrarian:latest | Simple |

### Subtitles
| Service | Port | Image | Present in |
|---------|------|-------|------------|
| Bazarr | 6767 | ghcr.io/hotio/bazarr:latest | VPN, Simple |

### Streaming
| Service | Port | Image | Present in |
|---------|------|-------|------------|
| Jellyfin | 8096 | lscr.io/linuxserver/jellyfin:latest | VPN, Simple |

### Request Management
| Service | Port | Image | Present in |
|---------|------|-------|------------|
| Seerr | 5055 | ghcr.io/seerr-team/seerr:v3.4.1 | VPN, Simple |

### Transcoding
| Service | Port | Image | Present in |
|---------|------|-------|------------|
| FileFlows | 19200 | revenz/fileflows:latest | VPN |

### VPN
| Service | Ports | Image | Present in |
|---------|-------|-------|------------|
| Gluetun | Multiple | qmcgaw/gluetun:latest | VPN |

---

## 📊 Data Structures

### Service
```go
type Service struct {
    ID            string
    Name          string
    Category      ServiceCategory
    Description   string
    Image         string
    ContainerName string
    Ports         []PortMapping
    Volumes       []VolumeMapping
    Environment   []string
    Devices       []string
    CapAdd        []string
    Network       NetworkConfig
    Restart       string
    SupportsVPN   bool
    RequiresVPN   bool
    Dependencies  []string
    Optional      bool
}
```

### Profile

```go
type Profile struct {
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Version     string
    VPN         VPNConfig
    Services    []string
    Environment map[string]string
    OutputDir   string
}
```

---

## 🎨 Usage Flow

### 1. Complete Interactive Mode
```bash
./corsarr generate

# Step 1: VPN Configuration
? Do you want to use VPN (Gluetun)? (y/N) › No

# Step 2: Service Selection
? Select the services you want to use:
  Download Managers:
    ☑ qBittorrent

  Indexers:
    ☑ Prowlarr
    ☐ FlareSolverr (requires VPN)

  Media Management:
    ☑ Sonarr (TV Shows)
    ☑ Radarr (Movies)
    ☐ Lidarr (Music)
    ☐ LazyLibrarian (Books)

  Subtitles:
    ☑ Bazarr

  Streaming:
    ☑ Jellyfin

  Request Management:
    ☐ Seerr (requires Jellyfin)

  Transcoding:
    ☐ FileFlows

# Step 3: Basic Configuration
? Base path (ARRPATH): › /home/user/media/
? Timezone (TZ): › America/Sao_Paulo
? User ID (PUID): › 1000
? Group ID (PGID): › 1000
? UMASK: › 002

# Step 4: Validation
✓ Configuration validated successfully!
✓ 6 services will be configured
✓ Mode: WITHOUT VPN
✓ No port conflicts detected

# Step 5: Confirmation
? Confirm file generation? (Y/n) › Yes

# Step 6: Generation
✓ Backup created: docker-compose.yml.backup
✓ Backup created: .env.backup
✓ docker-compose.yml created successfully
✓ .env created successfully

Files created in: /home/user/media/

To start services, run:
  cd /home/user/media/
  docker compose up -d

To check logs:
  docker compose logs -f
```

### 2. Using Existing Profile
```bash
./corsarr generate --profile basic-no-vpn

✓ Profile 'basic-no-vpn' loaded
✓ docker-compose.yml created
✓ .env created
```

### 3. Preview Without Generating
```bash
./corsarr preview

# Shows the content of files that would be generated
```

### 4. Non-Interactive Mode (CI/CD)
```bash
./corsarr generate --no-interactive \
  --services "prowlarr,radarr,sonarr,jellyfin,qbittorrent" \
  --arr-path "/home/user/media" \
  --timezone "America/Sao_Paulo" \
  --puid "1000" \
  --pgid "1000"
```

---

## 🔍 Implemented Validations

### 1. Port Conflicts
- Checks for duplicate ports between services
- Warns about ports already in use on the system (optional)

### 2. Service Dependencies
```
Seerr → requires Jellyfin
FlareSolverr → useful with Prowlarr
Bazarr → requires Sonarr OR Radarr
FileFlows → requires Jellyfin
```

### 3. Path Validation
- Verifies if ARRPATH exists or can be created
- Validates write permissions
- Checks available space (warning if < 10GB)

### 4. VPN Validation
- If VPN selected, validates required credentials
- Checks Wireguard key format
- Validates provider supported by Gluetun

### 5. Environment Validation
- Checks if Docker is installed
- Checks if Docker Compose is installed
- Validates minimum Docker version

---

## 🎁 Additional Features

### 1. Profile System
```bash
# Save current configuration
./corsarr profile save complete

# List profiles
./corsarr profile list

# Load profile
./corsarr generate --profile complete

# Remove profile
./corsarr profile delete complete

# Export profile
./corsarr profile export complete backup.json

# Import profile
./corsarr profile import backup.json --name restored
```

### 2. Automatic Backup
- Before generating new files, backs up existing ones
- Format: `docker-compose.yml.backup.TIMESTAMP`
- Keeps last 5 backups (configurable)

### 3. Dry-Run Mode
```bash
./corsarr generate --dry-run
# Only shows what would be done, without creating files
```

### 4. Health Check
```bash
./corsarr health
# Checks if all configured services are running
# Shows status of each container
```

### 5. Ports Check
```bash
./corsarr check-ports
# Checks which ports are in use on the system
# Suggests alternative ports if there's a conflict
```

---

## 📦 Direct Go Dependencies

```go
require (
    github.com/charmbracelet/huh v0.8.0 // Interactive prompts
    github.com/nicksnyder/go-i18n/v2 v2.4.0 // Internationalization
    github.com/spf13/cobra v1.8.0 // CLI framework
    golang.org/x/text v0.31.0 // Locale support
    gopkg.in/yaml.v3 v3.0.1 // YAML parsing
)
```

---

## 🔐 Security

### Implementation Analysis ✅

#### 1. Never log passwords or keys ✅
- [x] **Implemented**: Passwords use `EchoMode(huh.EchoModePassword)` (internal/prompts/config.go:47)
- [x] **Verified**: No `fmt.Print` of passwords/keys found in code
- [x] **Profiles**: Passwords stored in profiles (JSON/YAML) with `omitempty` tag
- [x] **Recommendation**: Consider encryption for profiles in future versions

#### 2. .env file with appropriate permissions ✅
- [x] **Implemented**: `.env` created with `0600` (internal/generator/env.go:68)
- [x] **Backups**: Backup files also use `0600`
- [x] **Test**: Automated test validates permissions

#### 3. Validate user inputs ✅
- [x] **Path Validation**: `internal/validator/path.go` validates:
  - Empty paths
  - Directory existence
  - Write permissions
  - Available disk space
- [x] **Port Validation**: `internal/validator/ports.go` detects conflicts
- [x] **Dependencies**: `internal/validator/dependencies.go` validates services
- [x] **Docker**: `internal/validator/docker.go` checks installation

#### 4. Sanitize paths ✅
- [x] **filepath.Join**: Used everywhere for path construction
- [x] **filepath.Clean**: Implicit in `filepath.Join` usage
- [x] **MkdirAll**: Uses `0755` for secure directory permissions
- [x] **Path traversal**: No insecure concatenation found

#### 5. Don't execute shell commands with user input ✅
- [x] **exec.Command**: Always uses fixed arguments, never user input
- [x] **Docker commands**: Paths passed as separate arguments
- [x] **No shell injection**: No use of `bash -c` or command concatenation

### Recommended Improvements (Post-v1.0.0)

1. **Profile encryption** (MEDIUM PRIORITY):
   - Encrypt passwords in saved profiles
   - Use OS keyring

2. **Security audit** (LOW PRIORITY):
   - Add automated security tests
   - Dependency vulnerability scanning
   - CodeQL analysis in GitHub Actions

---

## 📚 References

- [Docker Compose Specification](https://docs.docker.com/compose/compose-file/)
- [Gluetun Documentation](https://github.com/qdm12/gluetun-wiki)
- [LinuxServer.io Images](https://fleet.linuxserver.io/)
- [Cobra CLI](https://cobra.dev/)
- [Huh Forms](https://github.com/charmbracelet/huh)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)

---

### 🛠️ Tech Stack

- **Language**: Go 1.24.2
- **CLI Framework**: Cobra v1.8.0
- **TUI**: Huh v0.8.0 + Bubble Tea v1.3.10
- **Testing**: Standard Go testing
- **YAML**: gopkg.in/yaml.v3
- **Docker Integration**: os/exec (health, check-ports)
