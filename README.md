# Corsarr 🏴‍☠️

> Navigate the high seas of media automation

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
- 🎫 **Jellyseerr** - Request management interface
- 🔓 **FlareSolverr** - Bypass Cloudflare restrictions
- 📹 **FileFlows** - Transcode and optimize media
- 🔒 **Gluetun** - VPN client (optional)

**The CLI handles all the complexity** - service dependencies, network configuration, environment variables, port management, and more.

## ⚡ Quick Start (For Beginners)

### Prerequisites

- **Docker & Docker Compose v2+** - [Install here](https://docs.docker.com/compose/install/)
- Linux, macOS, or Windows with WSL2

### Installation

**Download the latest release for your platform:**

**Linux (AMD64):**

```bash
wget https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_Linux_x86_64.tar.gz
tar -xzf corsarr_Linux_x86_64.tar.gz
sudo mv corsarr /usr/local/bin/
```

**Linux (ARM64):**

```bash
wget https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_Linux_arm64.tar.gz
tar -xzf corsarr_Linux_arm64.tar.gz
sudo mv corsarr /usr/local/bin/
```

**macOS (Intel):**

```bash
wget https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_Darwin_x86_64.tar.gz
tar -xzf corsarr_Darwin_x86_64.tar.gz
sudo mv corsarr /usr/local/bin/
```

**macOS (Apple Silicon):**

```bash
wget https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_Darwin_arm64.tar.gz
tar -xzf corsarr_Darwin_arm64.tar.gz
sudo mv corsarr /usr/local/bin/
```

**Windows (PowerShell):**

```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_Windows_x86_64.zip" -OutFile "corsarr.zip"
Expand-Archive -Path "corsarr.zip" -DestinationPath "C:\Program Files\corsarr"

# Add to PATH (permanent)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\corsarr", [EnvironmentVariableTarget]::Machine)
```

Or download manually from [releases](https://github.com/woliveiras/corsarr/releases/latest) and extract to a folder in your PATH.

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
- **Jellyseerr** (Request content): http://localhost:5055
- **Radarr** (Movies): http://localhost:7878
- **Sonarr** (TV Shows): http://localhost:8989
- **Prowlarr** (Search): http://localhost:9696

---

## 🎯 Usage

### For Beginners: Interactive Mode

The CLI will ask you questions and generate everything automatically:

```bash
corsarr generate
```

**You'll be asked about**:

1. **Language** - Choose your preferred language
2. **VPN** - Do you want to route traffic through a VPN?
3. **Services** - Select which services you need
4. **Configuration** - Set paths, timezone, and user IDs
5. **VPN Details** - If enabled, configure your VPN provider

**The CLI creates**:
- `docker-compose.yml` - Complete service configuration
- `.env` - All environment variables

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

### 6. Jellyseerr (Requests)

Access `http://localhost:5055`

1. **Sign in with Jellyfin** account
2. **Connect to Radarr/Sonarr**: Settings → Services
3. **Allow users to request** content

---

## 🆘 Troubleshooting

### Service Can't Access Files

**Problem**: Permission denied errors

```bash
# Fix ownership
sudo chown -R 1000:1000 /path/to/media

# Verify PUID/PGID match
id $(whoami)
```

### Port Already in Use

**Problem**: Port conflict errors

```bash
# Check which ports are in use
corsarr check-ports --suggest

# Or check manually
docker ps
netstat -tuln | grep LISTEN
```

### VPN Not Working

**Problem**: Gluetun keeps restarting

```bash
# Check Gluetun logs
docker compose logs gluetun

# Test VPN connection
docker exec gluetun curl ifconfig.me
```

**Common fixes**:
- Verify VPN credentials in `.env`
- Regenerate WireGuard keys from provider
- Check provider is supported by Gluetun

### Container Won't Start

**Problem**: Service keeps restarting

```bash
# Check health status
corsarr health --detailed

# View service logs
docker compose logs [service_name]

# Check for errors
docker compose ps
```

### Can't Connect to Other Services

**Problem**: Radarr can't reach Prowlarr

**Solution**: Use container names, not localhost
- ✅ Correct: `http://prowlarr:9696`
- ❌ Wrong: `http://localhost:9696`

### High CPU Usage

**Problem**: Container using too much CPU

```bash
# Check resource usage
docker stats
corsarr health --detailed
```

**Common causes**:
- Jellyfin transcoding (normal during playback)
- Radarr/Sonarr scanning library (temporary)
- qBittorrent seeding (limit in settings)

### Database Locked

**Problem**: "Database is locked" errors

```bash
# Stop the affected service
docker compose stop sonarr

# Backup database
cp config/sonarr/*.db config/sonarr/backup/

# Restart service
docker compose start sonarr
```

### Need More Help?

**Collect diagnostic information**:

```bash
# System information
uname -a
docker --version
docker compose version

# Health report
corsarr health --detailed > health-report.txt

# Service logs
docker compose logs --tail=100 > logs.txt
```

**Get help**:

- Check [GitHub Issues](https://github.com/woliveiras/corsarr/issues)

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
cd /tmp
wget https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_Linux_x86_64.tar.gz
tar -xzf corsarr_Linux_x86_64.tar.gz
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
