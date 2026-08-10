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

`internal/orchestrator.Installer` owns the first transactional application
workflow: resolve and validate the approved spec, ensure the network, pull,
create, start, and inspect. A failure after container creation removes only that
owned incomplete container using a non-canceled cleanup context. Bind-mounted
configuration and media are deliberately outside cleanup. The orchestrator is
covered through the runtime interface and reached from Wails only through the
bounded `InstallSelectedApplications` intent.

Installation is reconciliatory: a matching running container is reused, and a
matching stopped container is started. A differently pinned image is never
replaced implicitly; that case is routed to the future update/rollback flow.
`internal/application.InstallationService` enforces current consent, prepares
the reviewed layout, orders dependencies before consumers, and returns a
structured per-application result while preserving already completed apps.
`internal/provisioning.HTTPReadiness` then probes only the catalog-resolved
loopback URL, without credentials or redirects, until the application accepts
HTTP or a bounded timeout expires. A newly created container that never becomes
ready is removed while its bind-mounted data remains available for diagnosis;
an existing container is never removed by this check.

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
Radarr, Sonarr, and Lidarr categories. The HTTP client is loopback-only,
redirect-free, cookie-scoped, and response-bounded.

The Wails surface can report only whether qBittorrent access is available and
the non-secret username. Password retrieval remains in Go: an explicit
`CopyQBittorrentPassword` intent writes it directly to the native clipboard and
returns no secret to TypeScript.

`internal/provisioning.ARRDownloadClientProvisioner` loads both secrets only in
Go and asks `ARRClient` to reconcile the reserved `qBittorrent (Corsarr)`
provider. The client first fetches the app's official live provider schema,
requires every expected qBittorrent field, fills only the internal network
alias, port, stored credentials, and app-specific category, then creates or
updates only the exact reserved provider name. Other user providers are never
selected. Radarr and Sonarr use API v3; Lidarr uses API v1.

`internal/provisioning.ProwlarrProvisioner` runs when each supported target Arr
app becomes ready. It reads the Prowlarr and target keys from their fixed config
files and reconciles only `<App> (Corsarr)` through Prowlarr API v1. The client
starts from the live application schema, retains its default category lists,
sets full sync plus the internal `prowlarr` and target network URLs, and relies
on Prowlarr's provider create/update path to validate connectivity. User-created
Prowlarr applications remain untouched.

`internal/provisioning.BazarrProvisioner` runs only after its explicit Radarr
and Sonarr dependencies are ready. It reads Bazarr's generated API key only
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

`internal/provisioning.SeerrProvisioner` runs after its Jellyfin, Radarr, and
Sonarr dependencies. It loads their credentials only in Go, authenticates Seerr
through the official Jellyfin login route, and retains the resulting HTTP-only
session in a private cookie jar. `SeerrClient` discovers/enables Jellyfin
libraries, tests both Arr connections to obtain live profiles and root folders,
prefers the `Any` quality profile (or the lowest returned ID), and reconciles
only the reserved `<App> (Corsarr)` entries with safe defaults. It initializes
Seerr only after every connection succeeds. The client is loopback-only,
proxy-free, redirect-free, and response-bounded.

`internal/application.ManagementService` translates runtime inspection into
`not_installed`, `running`, `stopped`, or `attention` for every catalog app. Its
start, stop, restart, and remove intents validate the catalog ID before reaching
the runtime. Remove deletes only the labeled container; the bind-mounted config
and shared media tree are outside its authority.

`internal/application.DataManagementService` is a distinct destructive-action
boundary. It requires the catalog application container to be absent, reloads
the persisted storage location, and delegates only that application's config
directory to `internal/storage.ApplicationDataManager`. The storage manager
refuses unsafe IDs and symlink targets, then atomically moves configuration into
the private `<selected>/Corsarr/trash/config/<app>/` tree. It never receives or
targets the shared media and downloads paths.

`internal/storage` inspects only a directory returned by the native Wails folder
picker. It verifies that the path already exists and is a directory, creates
temporary write and hardlink probes, reports available space, and removes all
probe artifacts. A ready selection is persisted through
`internal/application.SetupService` and `internal/state.FileStore`; rejected or
canceled selections are not persisted. The state file contains only the storage
path and approved application IDs, uses the operating system's user
configuration directory, and is written with private file permissions where the
platform supports them.

The state schema also records the accepted runtime-terms version and UTC
timestamp. Schema 1 setup files migrate to schema 2 without inventing consent.
Folder preparation remains available after storage/application review, but the
backend reports installation authority only when the current terms version was
explicitly accepted. A future terms version therefore requires new consent.

Application selection is validated against the presentation-safe catalog.
Required catalog dependencies are included recursively and the deterministic
selection is persisted. The frontend cannot provide an arbitrary directory name
to the layout operation: it can only request preparation from the persisted,
reviewed setup.

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
