# Corsarr CLI Architecture

> 🏴‍☠️ Navigate the high seas of media automation

> This document describes the current CLI and Docker Compose generator. The
> accepted, not-yet-implemented Corsarr Desktop direction is documented in
> [RFC 0001](rfcs/0001-corsarr-desktop.md). Durable desktop technology and
> runtime choices are recorded under [decisions](decisions/).

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
| Jellyseerr | 5055 | fallenbagel/jellyseerr:latest | VPN, Simple |

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
    ☐ Jellyseerr (requires Jellyfin)

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
Jellyseerr → requires Jellyfin
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
