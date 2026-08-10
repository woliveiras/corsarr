# Corsarr 🏴

> 🏴‍☠️ Navigate the high seas of media automation

<p align="center">
  <img src="assets/corsarr-logo-transparent.png" alt="Corsarr Logo" width="300"/>
</p>

The easiest way to set up and manage your complete media automation stack with Docker Compose.

**No configuration files to edit. No YAML to learn. Just answer a few questions.**

## 📖 What is Corsarr?

Corsarr is a CLI tool that generates complete Docker Compose configurations for your media automation stack. It includes:

- 🔍 **Prowlarr** - Search for torrents across multiple indexers
- 🎬 **Radarr** - Automatically download and organize movies
- 📺 **Sonarr** - Manage TV show downloads and library
- 🎵 **Lidarr** - Music collection manager
- 📚 **LazyLibrarian** - Book manager
- 💬 **Bazarr** - Automatic subtitle downloads
- ⬇️ **qBittorrent** - Torrent client for downloads
- 🎭 **Jellyfin** - Stream your media library
- 🎫 **Seerr** - Request management interface
- 🔓 **FlareSolverr** - Bypass Cloudflare restrictions
- 📹 **FileFlows** - Transcode and optimize media
- 🔒 **Gluetun** - VPN client (optional)

**The CLI handles all the complexity** - service dependencies, network configuration, environment variables, port management, and more.

## 🧭 Project Direction

Corsarr currently ships as the CLI documented below. Development has started on
Corsarr Desktop: a focused visual application for non-technical
users to install and operate the media stack on their personal computer while
keeping Docker, Podman, WSL, and other runtime details out of the primary user
experience.

- [Documentation index](docs/README.md)
- [Corsarr Desktop RFC](docs/rfcs/0001-corsarr-desktop.md)
- [Accepted architecture decisions](docs/decisions/)

The current development build provides a native Wails shell backed by the
existing Go service catalog. It lists user-facing applications, opens their
allowlisted local web interfaces, detects the local Docker runtime without
mutating it, and validates a user-selected storage folder for free space,
writing, and hardlink support. It also persists the reviewed folder and
application selection, includes catalog dependencies automatically, and creates
an idempotent `Corsarr/` media/configuration tree only after explicit
confirmation. A reviewed one-click preset selects the complete movie/TV stack
while leaving music, books, and transcoding optional. Existing installed apps
remain part of desired state, and removal is blocked while an installed app
still depends on the target. The development build can now record explicit
consent, prepare or start the supported runtime, and install the selected,
digest-pinned containers through the ownership-safe runtime boundary. Installation waits for
each allowlisted local web endpoint to become responsive before moving to its
dependants. Radarr, Sonarr, and Lidarr then receive their approved library root
folders idempotently through local authenticated APIs. qBittorrent's temporary
administrator credential is replaced with a generated `corsarr` credential
stored in the macOS Keychain; approved download paths and per-app categories are
reconciled, and the user can copy the password explicitly from the desktop UI.
Radarr, Sonarr, and Lidarr receive a dedicated `qBittorrent (Corsarr)` download
client built from each app's live provider schema, without modifying providers
created by the user. Prowlarr also receives reserved full-sync connections to
those three apps using their internal network URLs and generated API keys.
Bazarr requires Radarr and Sonarr, reads their generated keys only in Go, and
connects to both through Bazarr's authenticated settings API. On macOS, the
desktop can install or start its pinned, verified runtime after explicit
consent; Windows and Linux onboarding remain pending. Jellyfin's first-run
wizard is automated with a generated `corsarr` administrator stored in the
macOS Keychain, remote access disabled by default, and reserved movie, TV, and
music libraries. Installed
containers expose status plus open, start, stop, restart, and removal actions
that preserve data. Application configuration can be removed separately after
its container is gone; Corsarr shows its approximate size, moves it into a
private recoverable trash, and never targets the shared media or downloads tree.
The prerequisites and usage
instructions below describe the current CLI release.

The desktop can export a user-selected JSON diagnostic snapshot containing
versions, runtime and application health, and storage capabilities. The file is
written privately and atomically, redacts credential-shaped values, and never
includes application logs, cookies, passwords, API keys, or runtime sockets.
On macOS 13+, an explicit optional setting registers Corsarr through Apple's
Service Management framework. At login, Corsarr can start an already installed
runtime and existing owned containers without installing new applications;
pending approval remains visible in both Corsarr and System Settings.
Before preparing the runtime, the desktop performs a read-only Mac preflight:
supported Intel/Apple Silicon architecture, macOS 14 or newer, at least 4 GiB
of RAM, and at least 4 GiB free on the runtime-cache volume. All measured facts
remain visible under technical details, and a failed requirement blocks the
mutating preparation intent in both the UI and Go backend.
When Jellyfin is selected, a separate pre-install choice can expose only its
HTTP service to the local network for TVs and mobile clients. All administrative
applications remain bound to this computer by default. While Jellyfin is
running, the desktop lists a private IPv4 address that can be copied for a TV
or mobile device on the same network; known virtual runtime and tunnel
interfaces are not advertised.

The selected storage filesystem must be writable, measurable, and have at
least 10 GiB free. Hardlink support is reported as an efficiency warning rather
than a compatibility failure. Corsarr repeats this check immediately before
folder preparation, installation, and application update; update also repeats
the Mac runtime-cache free-space preflight. During installation,
the UI reports the current application and whether it is being started or configured without
exposing runtime logs or credentials.
Completed-operation payloads likewise expose only bounded success, failure,
rollback, and attention flags; private backup/trash paths, checksums, raw runtime
status, and backend error text remain in Go. Failed installation, provisioning,
and update results include a stable diagnostic code plus a plain-language next
action, available through expandable technical details without exposing the raw
backend failure. Application cards use the same bounded issue contract when
their state cannot be inspected. Explicit diagnostic export still includes a
separately redacted technical detail for support.
The Go desktop boundary also permits only one setup, runtime, or application-data
change at a time, so update, lifecycle, removal, onboarding, and login recovery
cannot race each other even if separate interface controls are activated.

Seerr signs in through that local Jellyfin administrator, enables the discovered
Jellyfin libraries, tests the internal Radarr/Sonarr connections, and reconciles
only the reserved `Radarr (Corsarr)` and `Sonarr (Corsarr)` instances before
marking its first-run setup complete.

### Run Corsarr Desktop from source

Development currently targets macOS on Apple Silicon first and requires Go
1.25+, Node.js 22+, pnpm 11.1.3, Xcode command-line tooling, and Wails v2:

```bash
corepack enable
corepack install
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
cd desktop
$(go env GOPATH)/bin/wails dev
```

Build the local macOS application with:

```bash
cd desktop
$(go env GOPATH)/bin/wails build
open build/bin/Corsarr.app
```

Opt-in Docker contracts can use an immutable image that is already present
locally. They create only labeled contract containers and the owned `corsarr`
network, exercise lifecycle/log inspection plus transactional installation and
HTTP readiness, and remove their resources afterward without pulling an image.
Run the packages serially because both deliberately exercise the same network:

```bash
CORSARR_DOCKER_CONTRACT_IMAGE='repository/name@sha256:<digest>' \
  go test -p 1 ./internal/runtime ./internal/orchestrator -run 'Real.*Contract' -v
```

To include real update and rollback replacement, also provide a different
immutable local image through `CORSARR_DOCKER_ROLLBACK_IMAGE`. Both references
are inspected locally and never pulled by the contract.

Catalog maintainers can query every approved immutable OCI index without
pulling layers and require both Linux AMD64 and ARM64 variants:

```bash
CORSARR_VERIFY_REMOTE_MANIFESTS=1 \
  go test ./internal/catalog -run '^TestRuntimeCatalogRemotePlatformContract$' -v
```

Managed containers carry a SHA-256 fingerprint of their image-independent
runtime contract. Automatic updates stop before backup, pull, or replacement
when the installed ports, mounts, init flag, or environment no longer match the
approved contract; such changes require a separately reviewed migration.
Idempotent installation applies the same check before reusing an existing
container, so a matching image cannot conceal stale mounts or network exposure.

This development build is self-signed and is not a published Corsarr release.
Run `pnpm run quality` from `desktop/frontend` to check formatting, lint rules,
and TypeScript types with Biome and `tsc`.

Pull requests and `main` run Go dependency verification, vet, race-enabled
tests with a retained coverage artifact, golangci-lint v2.11.1, the frontend
quality/build commands, and CGO-free CLI builds for the supported OS/CPU
matrix. CI reads the Go version from `go.mod`, installs pnpm from the pinned
`packageManager` field, grants only repository read access, and pins every
third-party action to a reviewed commit SHA.

Tag-triggered CLI releases install a pinned Syft binary and use GoReleaser to
attach an SPDX JSON software bill of materials for every archive and for the
source artifact. These SBOMs describe software actually packaged by the
release; media applications and container images downloaded later remain
identified separately in the desktop credits screen.

## ⚡ Quick Start

### Prerequisites

- **Docker & Docker Compose v2+** - [Install here](https://docs.docker.com/compose/install/)
- Linux, macOS, or Windows with WSL2

### Installation

**Download the latest release for your platform:**

<details>
<summary><strong>Linux (AMD64)</strong></summary>

```bash
curl -sL https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_linux_amd64.tar.gz | tar xz
sudo mv corsarr /usr/local/bin/
```

</details>

<details>
<summary><strong>Linux (ARM64)</strong></summary>

```bash
curl -sL https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_linux_arm64.tar.gz | tar xz
sudo mv corsarr /usr/local/bin/
```

</details>

<details>
<summary><strong>macOS (Intel)</strong></summary>

```bash
curl -sL https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_darwin_amd64.tar.gz | tar xz
sudo mv corsarr /usr/local/bin/
```

</details>

<details>
<summary><strong>macOS (Apple Silicon)</strong></summary>

```bash
curl -sL https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_darwin_arm64.tar.gz | tar xz
sudo mv corsarr /usr/local/bin/
```

</details>

<details>
<summary><strong>Windows (PowerShell)</strong></summary>

```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_windows_amd64.zip" -OutFile "corsarr.zip"
Expand-Archive -Path "corsarr.zip" -DestinationPath "C:\Program Files\corsarr"

# Add to PATH (permanent)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\corsarr", [EnvironmentVariableTarget]::Machine)
```

</details>


Or download manually from [releases](https://github.com/woliveiras/corsarr/releases/latest) and extract to a folder in your PATH.

---

## ✨ Key Features

- ✅ **Interactive CLI** - No configuration files to edit
- ✅ **Multi-language support** - English, Portuguese, Spanish
- ✅ **VPN Support** - Route traffic through Gluetun with WireGuard/OpenVPN
- ✅ **Automatic directory creation** - Sets up all needed folders automatically
- ✅ **Port conflict detection** - Validates ports before generating files
- ✅ **Profile management** - Save and reuse configurations
- ✅ **Non-interactive mode** - Perfect for automation and CI/CD
- ✅ **Cross-platform** - Linux, macOS, Windows (AMD64 and ARM64)
- ✅ **Health monitoring** - Check service status and resource usage
- ✅ **Dry-run mode** - Preview configuration before creation

---

### Usage

```bash
# 1. Generate your stack (interactive)
corsarr generate

# 2. Start everything
docker compose up -d
```

**That's it!** Your media automation stack is now running. 🎉

Access your services:

- **Jellyfin** (Watch movies/TV): http://localhost:8096
- **Seerr** (Request content): http://localhost:5055
- **Radarr** (Movies): http://localhost:7878
- **Sonarr** (TV Shows): http://localhost:8989
- **Prowlarr** (Search): http://localhost:9696

---

## 🎯 Usage

### Interactive Mode

The CLI will ask you questions and generate everything automatically:

```bash
corsarr generate
```

**You'll be asked about**:

1. **Language** - Choose your preferred language
2. **VPN** - Do you want to route traffic through a VPN?
3. **Services** - Select which services you need
4. **Configuration** - Set paths, timezone, and user IDs
5. **Output Directory** - Where to generate files (optional)
6. **VPN Details** - If enabled, configure your VPN provider

**The CLI creates**:

- `docker-compose.yml` - Complete service configuration
- `.env` - All environment variables
- **All necessary directories** - Config and data folders for each service

**Then start your stack**:
```bash
docker compose up -d
```

### Useful Commands

**Check if everything is healthy**:
```bash
corsarr health
corsarr health --detailed  # With CPU/memory stats
```

**Check for port conflicts**:
```bash
corsarr check-ports
corsarr check-ports --suggest  # Get alternative ports
```

**Preview configuration without creating files**:
```bash
corsarr preview
```

**Save your setup for later**:
```bash
corsarr generate --save-profile --save-as my-setup
```

**Reuse a saved configuration**:
```bash
corsarr generate --profile my-setup
```

---

## 🚀 Advanced Usage

### Generate with Custom Options

**Specify output directory**:
```bash
corsarr generate --output ~/my-media-stack
```

**Enable VPN mode directly**:
```bash
corsarr generate --vpn
```

**Preview without creating files**:
```bash
corsarr generate --dry-run
```

**Use a specific language**:
```bash
corsarr generate --language pt-br  # Portuguese
corsarr generate --language es     # Spanish
```

### Profile Management

Profiles let you save and reuse configurations:

**List all profiles**:
```bash
corsarr profile list
```

**Load a profile to see details**:
```bash
corsarr profile load my-setup
```

**Delete a profile**:
```bash
corsarr profile delete old-setup
```

**Export profile to share**:
```bash
corsarr profile export my-setup backup.json
```

**Import profile from file**:
```bash
corsarr profile import backup.json
corsarr profile import backup.json --name new-name
```

### Non-Interactive Mode (CI/CD)

For scripts, automation, and continuous deployment:

```bash
corsarr generate --no-interactive \
  --services "prowlarr,radarr,sonarr,jellyfin,qbittorrent" \
  --arr-path "/home/user/media" \
  --timezone "America/Sao_Paulo" \
  --puid "1000" \
  --pgid "1000" \
  --output ./stack
```

**With VPN**:

```bash
corsarr generate --no-interactive \
  --vpn \
  --vpn-provider protonvpn \
  --vpn-password "your-wireguard-key" \
  --services "radarr,sonarr,qbittorrent" \
  --arr-path "/media" \
  --timezone "UTC" \
  --puid "1000" \
  --pgid "1000"
```

**Using configuration file**:

```yaml
# config.yaml
services:
  - prowlarr
  - radarr
  - sonarr
  - jellyfin
  - qbittorrent
arr_path: /home/user/media
timezone: America/Sao_Paulo
puid: 1000
pgid: 1000
```

```bash
corsarr generate --config config.yaml --no-interactive
```

**All non-interactive flags**:

- `--no-interactive` - Skip all prompts
- `--services` - Comma-separated service list
- `--arr-path` - Base path for media library
- `--timezone` - Timezone (e.g., `America/Sao_Paulo`)
- `--puid` - User ID for file permissions
- `--pgid` - Group ID for file permissions
- `--umask` - File creation mask (default: `002`)
- `--project-name` - Docker Compose project name
- `--vpn` - Enable VPN mode
- `--vpn-provider` - VPN provider (required with `--vpn`)
- `--vpn-password` - WireGuard key or OpenVPN password
- `--vpn-type` - `wireguard` or `openvpn` (default: `wireguard`)
- `--config` - Load from YAML/JSON config file
- `--profile` - Load from saved profile

---

## ⚙️ Configuration

### Environment Variables

The CLI will prompt you for these values:

| Variable | Description | Example |
|----------|-------------|---------|
| `ARRPATH` | Base path for media library | `/home/user/media/` |
| `TZ` | Your timezone | `America/Sao_Paulo` |
| `PUID` | User ID (run `id -u`) | `1000` |
| `PGID` | Group ID (run `id -g`) | `1000` |
| `UMASK` | File creation mask | `002` |

**Finding your PUID/PGID**:

```bash
id $(whoami)
# Output: uid=1000(user) gid=1000(user)
```

### VPN Configuration

When VPN is enabled, you'll configure:

- **Provider** - nordvpn, protonvpn, expressvpn, etc. ([see all supported](https://github.com/qdm12/gluetun-wiki))
- **Type** - WireGuard (recommended) or OpenVPN
- **Credentials** - Username/password or WireGuard private key
- **Port Forwarding** - Enable for better torrent connectivity
- **DNS** - Custom DNS server (default: 1.1.1.1)

### Directory Structure

Corsarr automatically creates all necessary directories when generating files:

```
/your/media/path/
├── config/              # Service configurations
│   ├── radarr/
│   ├── sonarr/
│   ├── prowlarr/
│   ├── jellyfin/
│   └── ...
├── data/                # Media library
│   ├── movies/
│   ├── tvshows/
│   ├── music/
│   ├── books/
│   └── downloads/
└── backup/              # Automatic backups
```

If directories already exist, Corsarr will detect and reuse them without overwriting.

### Network Modes

**VPN Mode**: All traffic routes through Gluetun

```yaml
services:
  radarr:
    network_mode: "service:gluetun"
```

**Bridge Mode**: Direct network access (no VPN)

```yaml
services:
  radarr:
    networks:
      - media
```

---

## 🔧 Initial Service Configuration

After starting your stack, configure each service:

### 1. qBittorrent

Access `http://localhost:8080`

- **Default login**: `admin` / run `docker logs qbittorrent` for password
- **Set download path**: Tools → Options → Downloads → `/downloads`
- **Change password**: Tools → Options → Web UI → Authentication

### 2. Prowlarr

Access `http://localhost:9696`

1. **Add qBittorrent**: Settings → Download Clients → Add qBittorrent
   - Host: `qbittorrent`
   - Port: `8081`
   - Username/password from step 1

2. **Add indexers**: Indexers → Add Indexer
   - Choose your preferred torrent sites
   - Configure credentials

3. **Copy API Key**: Settings → General → Security → Copy API Key

### 3. Radarr (Movies) / Sonarr (TV Shows)

Access `http://localhost:7878` (Radarr) or `http://localhost:8989` (Sonarr)

1. **Add media folder**:
   - Settings → Media Management → Add Root Folder
   - Radarr: `/data/movies`
   - Sonarr: `/data/tvshows`

2. **Add qBittorrent**: Settings → Download Clients → Add qBittorrent
   - Host: `qbittorrent`
   - Port: `8081`

3. **Connect to Prowlarr**: Settings → Indexers → Add → Prowlarr
   - URL: `http://prowlarr:9696`
   - API Key: (from Prowlarr)

4. **Copy API Key**: Settings → General → Security → Copy API Key

### 4. Bazarr (Subtitles)

Access `http://localhost:6767`

1. **Add subtitle providers**: Settings → Providers
2. **Connect to Radarr**: Settings → Radarr
   - Address: `radarr`
   - Port: `7878`
   - API Key: (from Radarr)
3. **Connect to Sonarr**: Settings → Sonarr
   - Address: `sonarr`
   - Port: `8989`
   - API Key: (from Sonarr)

### 5. Jellyfin (Streaming)

Access `http://localhost:8096`

1. **Create admin account** during initial setup
2. **Add libraries**:
   - Movies: `/data/movies`
   - TV Shows: `/data/tvshows`
   - Music: `/data/music`
3. **Install Jellyfin apps** on your devices

### 6. Seerr (Requests)

Access `http://localhost:5055`

1. **Sign in with Jellyfin** account
2. **Connect to Radarr/Sonarr**: Settings → Services
3. **Allow users to request** content

---

## 🆘 Troubleshooting

The full troubleshooting guide is on a dedicated file: [Troubleshooting](docs/TROUBLESHOOTING.md)

---

## 🔒 Security Best Practices

- **Use VPN** - Route torrent traffic through a VPN
- **Change default passwords** - Update all service credentials
- **Restrict external access** - Use firewall rules to limit ports
- **Use reverse proxy** - Set up Nginx/Traefik with HTTPS for remote access
- **Keep updated** - Run `docker compose pull && docker compose up -d` regularly

---

## 📚 Example: Downloading Legal Content

Try downloading public domain content to test your setup:

1. **Open Radarr** (`http://localhost:7878`)
2. **Add movie**: Click "Add New Movie"
3. **Search**: Try "Night of the Living Dead (1968)"
4. **Monitor**: Select "Monitored"
5. **Search**: Click "Search" to find torrents

Watch it appear in qBittorrent, download, and show up in Jellyfin!

**More public domain movies**:

- Nosferatu (1922)
- The City of the Dead (1960)
- Plan 9 from Outer Space (1959)
- Find more at [JustWatch Public Domain](https://www.justwatch.com/us/provider/public-domain-movies)

---

## 📦 Backup and Restore

**Backup your configuration**:

```bash
# Backup config directory (includes databases)
cp -r config/ ~/corsarr-backup/

# Backup your media (optional, but recommended)
rsync -av data/ /path/to/external/drive/
```

**Restore from backup**:

```bash
# Restore configuration
cp -r ~/corsarr-backup/ config/

# Start services
docker compose up -d
```

**Automated backups**: Each service creates automatic backups in `config/[service]/Backups/`

---

## 🔄 Updating

**Update Corsarr CLI**:

Download the latest release from [GitHub Releases](https://github.com/woliveiras/corsarr/releases/latest) or use these commands:

```bash
# Linux/macOS - Download and replace
curl -sL https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_linux_amd64.tar.gz | tar xz
sudo mv corsarr /usr/local/bin/
```

**Update Docker containers**:

```bash
docker compose pull
docker compose up -d
```

---

---

## 📄 License

See [LICENSE](LICENSE) file.

---

## 🔗 Links

- **[Issue Tracker](https://github.com/woliveiras/corsarr/issues)** - Report bugs or request features
- **[Gluetun Wiki](https://github.com/qdm12/gluetun-wiki)** - VPN provider documentation

---

**Made with ❤️ by me for the community**
