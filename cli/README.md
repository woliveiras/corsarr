# Corsarr CLI 🏴‍☠️

> Navigate the high seas of media automation

CLI tool to easily configure and deploy your *arr stack (Radarr, Sonarr, Prowlarr, etc.) with Docker Compose.

## ✨ Features

- 🌍 **Multilingual**: English, Português Brasileiro, Español
- 🎯 **Interactive**: Modern TUI with intuitive prompts
- 🔧 **Flexible**: Choose only the services you need
- 🔒 **VPN Support**: Optional VPN integration with Gluetun
- 💾 **Profiles**: Save and reuse configurations
- ✅ **Validation**: Automatic port conflict and dependency checking
- 📦 **Export/Import**: Share profiles between systems
- 🔍 **Preview**: Dry-run mode before file creation

## 📦 Installation

### Prerequisites

- Docker and Docker Compose v2
- (Optional) Go 1.24.2+ to build from source

### Build from Source

```bash
git clone https://github.com/woliveiras/corsarr.git
cd corsarr/cli
go build -o corsarr .
sudo mv corsarr /usr/local/bin/  # Optional: install globally
```

### Quick Start

```bash
# Run directly without installing
go run main.go generate
```

## 🚀 Quick Start

### 1. Generate Your Stack

```bash
corsarr generate
```

This will:
1. Ask your preferred language
2. Prompt for VPN usage
3. Let you select services
4. Configure environment variables
5. Generate `docker-compose.yml` and `.env`

### 2. Start Your Services

```bash
docker compose up -d
```

### 3. Access Your Services

- **Prowlarr**: http://localhost:9696
- **Radarr**: http://localhost:7878
- **Sonarr**: http://localhost:8989
- **Jellyfin**: http://localhost:8096
- **qBittorrent**: http://localhost:8080
- *(ports depend on your configuration)*

## 📖 Usage Guide

### Generate Command

**Interactive mode** (recommended):
```bash
corsarr generate
```

**With custom output directory**:
```bash
corsarr generate --output /path/to/output
```

**Enable VPN mode**:
```bash
corsarr generate --vpn
```

**Preview without creating files**:
```bash
corsarr generate --dry-run
```

**Use saved profile**:
```bash
corsarr generate --profile my-setup
```

**Save configuration after generation**:
```bash
corsarr generate --save-profile --save-as my-setup
```

### Profile Commands

**List all profiles**:
```bash
corsarr profile list
```

**View profile details**:
```bash
corsarr profile load my-setup
```

**Save a new profile**:
```bash
corsarr profile save my-setup -d "My production setup"
```

**Delete a profile**:
```bash
corsarr profile delete my-setup
```

**Export profile to share**:
```bash
corsarr profile export my-setup backup.json
```

**Import profile from file**:
```bash
corsarr profile import backup.json
corsarr profile import backup.json --name new-name  # Rename on import
```

### Preview Command

**Preview configuration without generating**:
```bash
corsarr preview
corsarr preview --profile my-setup
```

## 🎮 Available Services

### Download Managers
- **qBittorrent** - BitTorrent client

### Indexers
- **Prowlarr** - Indexer manager for *arr apps
- **FlareSolverr** - Cloudflare bypass proxy

### Media Management
- **Sonarr** - TV show collection manager
- **Radarr** - Movie collection manager
- **Lidarr** - Music collection manager
- **LazyLibrarian** - Book collection manager

### Subtitles
- **Bazarr** - Subtitle downloader

### Streaming
- **Jellyfin** - Media streaming server

### Request Management
- **Jellyseerr** - Request management for movies and TV shows

### Transcoding
- **FileFlows** - Media transcoding and optimization

### VPN
- **Gluetun** - VPN client with multiple provider support

## ⚙️ Configuration

### Environment Variables

The CLI will prompt for these variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `ARRPATH` | Base path for media library | `/home/user/media/` |
| `TZ` | Timezone | `America/Sao_Paulo` |
| `PUID` | User ID | `1000` |
| `PGID` | Group ID | `1000` |
| `UMASK` | File creation mask | `002` |

### VPN Configuration

If VPN is enabled, you'll configure:

- **Provider**: nordvpn, protonvpn, expressvpn, etc.
- **Type**: WireGuard or OpenVPN
- **Credentials**: Username, password, or WireGuard keys
- **Port Forwarding**: Enable/disable
- **DNS**: Custom DNS server (default: 1.1.1.1)

### Network Modes

**VPN Mode**: All services route through Gluetun
```yaml
services:
  radarr:
    network_mode: "service:gluetun"
```

**Bridge Mode**: Each service on dedicated network
```yaml
services:
  radarr:
    networks:
      - media
```

## 📁 Generated Files

### docker-compose.yml

Complete Docker Compose configuration with:
- Selected services
- Network configuration (VPN or bridge)
- Volume mappings
- Environment variables
- Restart policies

### .env

Environment variables file with:
- Project name
- Paths and timezone
- User/group IDs
- VPN credentials (if enabled)

## 🔧 Advanced Usage

### Custom Output Directory

```bash
corsarr generate --output ~/docker/media-stack
cd ~/docker/media-stack
docker compose up -d
```

### Profile Workflow

```bash
# Create different profiles for different setups
corsarr generate --save-profile --save-as home-server
corsarr generate --save-profile --save-as vps-minimal

# Switch between them easily
corsarr generate --profile home-server
corsarr generate --profile vps-minimal
```

### Validation

Corsarr automatically validates:
- ✅ Port conflicts between services
- ✅ Service dependencies (e.g., Radarr requires Prowlarr)
- ✅ Path accessibility and permissions
- ✅ Docker and Docker Compose installation
- ⚠️ Available disk space warnings

## 🌍 Language Support

Change language anytime:
```bash
corsarr generate --language pt-br  # Português
corsarr generate --language en     # English
corsarr generate --language es     # Español
```

Or set via prompt on first run.

## 🆘 Troubleshooting

### Port Already in Use

If you see port conflict warnings:
1. Check running containers: `docker ps`
2. Modify ports in `docker-compose.yml`
3. Or stop conflicting services

### Permission Errors

Ensure `PUID` and `PGID` match your user:
```bash
id $(whoami)  # Check your UID and GID
```

### VPN Not Working

1. Verify VPN credentials in `.env`
2. Check Gluetun logs: `docker logs gluetun`
3. Ensure port forwarding is configured correctly

## 🤝 Contributing

Contributions are welcome! See [ARCHITECTURE.md](../docs/ARCHITECTURE.md) for technical documentation.

## 📄 License

See [LICENSE](../LICENSE) in the main repository.

## 🔗 Links

- [Main Repository](https://github.com/woliveiras/corsarr)
- [Technical Documentation](../docs/ARCHITECTURE.md)
- [Issue Tracker](https://github.com/woliveiras/corsarr/issues)

## License

See LICENSE file in the main repository.

## Links

- [Main Repository](https://github.com/woliveiras/corsarr)
- [Documentation](https://github.com/woliveiras/corsarr/tree/main/docs)
