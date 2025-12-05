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

### Health Command

**Check container health status**:
```bash
corsarr health
```

**Check with detailed CPU/memory stats**:
```bash
corsarr health --detailed
```

**Check specific directory**:
```bash
corsarr health --output /path/to/compose
```

This command will:

- Verify Docker is available
- Check all containers managed by your docker-compose.yml
- Show status (running/stopped/unhealthy)
- Display CPU and memory usage (with --detailed flag)
- Suggest actions for issues

### Check Ports Command

**Check for port conflicts**:
```bash
corsarr check-ports
```

**Get alternative port suggestions**:
```bash
corsarr check-ports --suggest
```

**Check specific directory**:
```bash
corsarr check-ports --output /path/to/compose
```

This command will:
- List all ports used by configured services
- Check if ports are available on your system
- Show which process is using conflicting ports
- Suggest alternative ports when conflicts are found

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

If you encounter port conflicts:

```bash
# Check which ports are in use
corsarr check-ports

# Get alternative port suggestions
corsarr check-ports --suggest

# Manually check running containers
docker ps
```

**Solutions**:
1. Stop conflicting containers: `docker compose down`
2. Modify ports in `docker-compose.yml`
3. Use suggested alternative ports from `check-ports`

### Container Not Starting

Use the health check command:

```bash
# Check all containers
corsarr health

# Get detailed info with CPU/memory
corsarr health --detailed
```

**Common issues**:
- **Unhealthy container**: Check logs with `docker compose logs [service]`
- **Stopped container**: Start with `docker compose up -d`
- **Missing dependencies**: Verify service dependencies are configured

### Permission Errors

**Problem**: Containers can't access files or create directories

**Solution**: Ensure `PUID` and `PGID` match your user:

```bash
# Check your user/group IDs
id $(whoami)

# Example output: uid=1000(user) gid=1000(user)
# Use these values for PUID and PGID
```

Then regenerate with correct IDs:
```bash
corsarr generate  # Enter correct PUID/PGID when prompted
```

### VPN Not Working

**Problem**: Services can't connect through VPN

**Troubleshooting steps**:

1. **Check Gluetun logs**:
   ```bash
   docker compose logs gluetun
   ```

2. **Verify VPN credentials** in `.env`:
   - `VPN_SERVICE_PROVIDER`
   - `WIREGUARD_PRIVATE_KEY` or `OPENVPN_USER`/`OPENVPN_PASSWORD`

3. **Test VPN connection**:
   ```bash
   docker exec gluetun curl ifconfig.me
   ```

4. **Common fixes**:
   - Regenerate WireGuard keys from your VPN provider
   - Enable port forwarding in `.env` if needed
   - Check firewall rules on your system

### Docker Compose Not Found

**Problem**: `docker compose` command not available

**Solution**:

```bash
# Check Docker Compose version
docker compose version

# If not installed, install Docker Compose v2
# Linux:
sudo apt-get update && sudo apt-get install docker-compose-plugin

# Or use docker-compose (v1) if available
docker-compose version
```

### Files Not Generated

**Problem**: `corsarr generate` completes but no files created

**Troubleshooting**:

1. **Check output directory permissions**:
   ```bash
   ls -la /path/to/output
   ```

2. **Use dry-run to see what would be generated**:
   ```bash
   corsarr generate --dry-run
   ```

3. **Verify there are no validation errors** in the output

4. **Try with explicit output path**:
   ```bash
   corsarr generate --output ~/corsarr-test
   ```

### Need More Help?

- **Check health status**: `corsarr health --detailed`
- **Validate configuration**: `corsarr preview`
- **Check ports**: `corsarr check-ports --suggest`
- **Review logs**: `docker compose logs -f`
- **Technical docs**: See [ARCHITECTURE.md](../docs/ARCHITECTURE.md)
- **Report issues**: [GitHub Issues](https://github.com/woliveiras/corsarr/issues)

## 🤝 Contributing

Contributions are welcome! See [ARCHITECTURE.md](../docs/ARCHITECTURE.md) for technical documentation.

## 📄 License

See [LICENSE](../LICENSE) in the main repository.

## 🔗 Links

- [Main Repository](https://github.com/woliveiras/corsarr)
- [Technical Documentation](../docs/ARCHITECTURE.md)
- [Issue Tracker](https://github.com/woliveiras/corsarr/issues)
