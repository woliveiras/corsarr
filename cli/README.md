# Corsarr CLI 🏴‍☠️

> Navigate the high seas of media automation

CLI tool to easily configure and deploy your *arr stack (Radarr, Sonarr, Prowlarr, etc.) with Docker Compose.

## Features

- 🌍 **Multilingual**: English, Português Brasileiro, Español
- 🎯 **Interactive**: Easy-to-use prompts for configuration
- 🔧 **Flexible**: Choose only the services you need
- 🔒 **VPN Support**: Optional VPN integration with Gluetun
- 💾 **Profiles**: Save and reuse configurations
- ✅ **Validation**: Automatic port conflict detection and dependency checking

## Installation

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose

### Build from Source

```bash
cd cli
go build -o corsarr
```

### Run Without Building

```bash
cd cli
go run main.go [command]
```

## Usage

### Generate Configuration

Interactive mode (recommended):
```bash
./corsarr generate
```

With flags:
```bash
./corsarr generate --vpn --output /path/to/output
```

Using a saved profile:
```bash
./corsarr generate --profile my-setup
```

Dry run (preview only):
```bash
./corsarr generate --dry-run
```

### Preview Configuration

```bash
./corsarr preview
./corsarr preview --profile my-setup
```

### Manage Profiles

List all profiles:
```bash
./corsarr profile list
```

Save current configuration:
```bash
./corsarr profile save my-setup
```

Delete a profile:
```bash
./corsarr profile delete my-setup
```

Export/Import profiles:
```bash
./corsarr profile export my-setup > my-setup.yaml
./corsarr profile import my-setup.yaml
```

## Available Services

### Download Managers
- **qBittorrent**: BitTorrent client

### Indexers
- **Prowlarr**: Indexer manager for *arr apps
- **FlareSolverr**: Cloudflare bypass proxy

### Media Management
- **Sonarr**: TV show collection manager
- **Radarr**: Movie collection manager
- **Lidarr**: Music collection manager
- **LazyLibrarian**: Book collection manager

### Subtitles
- **Bazarr**: Subtitle downloader

### Streaming
- **Jellyfin**: Media streaming server

### Request Management
- **Jellyseerr**: Request management for movies and TV shows

### Transcoding
- **FileFlows**: Media transcoding and optimization

### VPN
- **Gluetun**: VPN client with multiple provider support

## Configuration

The CLI will prompt you for:

- **Language**: Choose your preferred language
- **VPN**: Enable/disable VPN mode
- **Services**: Select which services to include
- **Environment Variables**:
  - `ARRPATH`: Base path for media library
  - `TZ`: Timezone
  - `PUID`: User ID
  - `PGID`: Group ID
  - `UMASK`: File creation mask
- **VPN Configuration** (if enabled):
  - Provider
  - VPN type
  - Wireguard keys
  - Port forwarding
  - DNS settings

## Generated Files

The CLI generates:

1. **docker-compose.yml**: Complete Docker Compose configuration
2. **.env**: Environment variables file

Both files are ready to use with `docker compose up -d`.

## Development Status

### ✅ Completed
- Project structure
- Internationalization system (EN, PT-BR, ES)
- Command structure (Cobra)
- Basic commands skeleton

### 🚧 In Progress
- Service definitions
- Template generation
- Interactive prompts
- Validation system

### 📋 Planned
- Profile management
- Health checks
- Port conflict detection
- Docker environment validation

## Project Structure

```
cli/
├── cmd/                    # CLI commands
├── internal/
│   ├── i18n/              # Internationalization
│   ├── services/          # Service definitions
│   ├── generator/         # File generators
│   ├── validator/         # Validation logic
│   ├── prompts/           # Interactive prompts
│   └── profile/           # Profile management
├── templates/
│   ├── docker-compose/    # Compose templates
│   └── services/          # Service definitions
├── locales/               # Translation files
│   ├── en.yaml
│   ├── pt-br.yaml
│   └── es.yaml
├── configs/
│   └── profiles/          # Saved profiles
├── go.mod
└── main.go
```

## Contributing

Contributions are welcome! Please see the main repository for guidelines.

## License

See LICENSE file in the main repository.

## Links

- [Main Repository](https://github.com/woliveiras/corsarr)
- [Documentation](https://github.com/woliveiras/corsarr/tree/main/docs)
