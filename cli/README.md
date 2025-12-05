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

### Non-Interactive Mode

**For scripts and CI/CD pipelines**, use non-interactive mode with all configuration via flags:

```bash
corsarr generate --no-interactive \
  --services "prowlarr,radarr,sonarr,jellyfin,qbittorrent" \
  --arr-path "/home/user/media" \
  --timezone "America/Sao_Paulo" \
  --puid "1000" \
  --pgid "1000" \
  --output ./docker-stack
```

**With VPN enabled**:
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
```bash
# Create config.yaml (see config.example.yaml)
corsarr generate --config config.yaml --no-interactive
```

**Using saved profile**:
```bash
corsarr generate --profile my-setup --no-interactive
```

**All non-interactive flags**:
- `--no-interactive` - Enable non-interactive mode
- `--services` - Comma-separated service list
- `--arr-path` - Base path for media library
- `--timezone` - Timezone (e.g., `America/Sao_Paulo`)
- `--puid` - User ID for permissions
- `--pgid` - Group ID for permissions
- `--umask` - File creation mask (default: `002`)
- `--project-name` - Docker Compose project name (default: `corsarr`)
- `--vpn` - Enable VPN mode
- `--vpn-provider` - VPN provider (required with `--vpn`)
- `--vpn-password` - WireGuard private key or OpenVPN password
- `--vpn-type` - VPN type: `wireguard` or `openvpn` (default: `wireguard`)
- `--config` - Load all config from YAML/JSON file
- `--profile` - Load config from saved profile

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

### CI/CD Integration

**GitHub Actions example**:
```yaml
name: Deploy Media Stack

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Corsarr
        run: |
          wget https://github.com/woliveiras/corsarr/releases/latest/download/corsarr-linux-amd64
          chmod +x corsarr-linux-amd64
          sudo mv corsarr-linux-amd64 /usr/local/bin/corsarr
      
      - name: Generate stack
        run: |
          corsarr generate --no-interactive \
            --services "prowlarr,radarr,sonarr" \
            --arr-path "${{ secrets.MEDIA_PATH }}" \
            --timezone "UTC" \
            --puid "1000" \
            --pgid "1000" \
            --output ./stack
      
      - name: Deploy
        run: |
          cd stack
          docker compose up -d
```

**Using config file in version control**:
```bash
# Keep config.yaml in your repo (without secrets)
# Use environment variables for sensitive data
export VPN_KEY="${WIREGUARD_KEY}"
envsubst < config.template.yaml > config.yaml
corsarr generate --config config.yaml --no-interactive
```

**Ansible playbook example**:
```yaml
- name: Deploy Corsarr stack
  hosts: media_servers
  tasks:
    - name: Generate docker-compose
      command: >
        corsarr generate --no-interactive
        --profile production
        --output /opt/media-stack
      
    - name: Start services
      community.docker.docker_compose:
        project_src: /opt/media-stack
        state: present
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

### Service Can't Access Media Files

**Problem**: "Permission denied" errors in service logs

**Symptoms**:
```
Error: unable to access /data/movies: permission denied
Failed to scan library: insufficient permissions
```

**Solutions**:

1. **Fix ownership** of media directories:
   ```bash
   sudo chown -R 1000:1000 /path/to/media
   ```

2. **Verify PUID/PGID** match your user:
   ```bash
   id $(whoami)
   # uid=1000(user) gid=1000(user)
   ```

3. **Check UMASK** in `.env`:
   ```bash
   UMASK=002  # Recommended for proper permissions
   ```

4. **Restart containers** after fixing permissions:
   ```bash
   docker compose restart
   ```

### Gluetun Keeps Restarting

**Problem**: VPN container in restart loop

**Check logs**:
```bash
docker compose logs gluetun | tail -50
```

**Common causes**:

1. **Invalid credentials**:
   ```
   Error: authentication failed
   ```
   → Regenerate WireGuard keys or check OpenVPN credentials

2. **Invalid provider**:
   ```
   Error: provider 'xyz' not supported
   ```
   → Check supported providers: https://github.com/qdm12/gluetun-wiki

3. **Network issues**:
   ```
   Error: cannot reach VPN server
   ```
   → Check firewall, DNS, or try different server location

4. **Missing environment variables**:
   ```
   Error: WIREGUARD_PRIVATE_KEY is not set
   ```
   → Verify all VPN variables in `.env`

**Fix**:
```bash
# Stop everything
docker compose down

# Edit .env with correct credentials
nano .env

# Start again
docker compose up -d
```

### Jellyfin Transcoding Not Working

**Problem**: Videos won't play or buffer constantly

**Solutions**:

1. **Enable hardware acceleration** (if available):
   ```bash
   # Check if /dev/dri exists
   ls -la /dev/dri
   ```

2. **Verify user is in video/render groups**:
   ```bash
   sudo usermod -aG video,render $USER
   # Logout and login again
   ```

3. **Check Jellyfin logs**:
   ```bash
   docker compose logs jellyfin | grep -i transcode
   ```

4. **Increase Jellyfin's RAM** in docker-compose.yml:
   ```yaml
   deploy:
     resources:
       limits:
         memory: 4G
   ```

### Prowlarr/Radarr/Sonarr Can't Connect

**Problem**: Services can't communicate with each other

**Symptoms**:
```
Unable to connect to Prowlarr
Connection refused to http://prowlarr:9696
```

**Solutions**:

1. **Check container names** match in docker-compose.yml:
   ```yaml
   services:
     prowlarr:
       container_name: prowlarr  # Must match connection URL
   ```

2. **Use container names** in service URLs (not localhost):
   ```
   ✅ Correct: http://prowlarr:9696
   ❌ Wrong:   http://localhost:9696
   ```

3. **Verify all containers are on same network**:
   ```bash
   docker inspect -f '{{.NetworkSettings.Networks}}' prowlarr radarr sonarr
   ```

4. **Check with curl inside container**:
   ```bash
   docker exec radarr curl -I http://prowlarr:9696
   ```

### High CPU Usage

**Problem**: Containers using excessive CPU

**Diagnosis**:

```bash
# Check resource usage
docker stats

# Identify the culprit
corsarr health --detailed
```

**Common causes**:

1. **Active transcoding** (Jellyfin/FileFlows):
   → Normal during video conversion, wait for completion

2. **Media scanning** (Radarr/Sonarr):
   → Normal after adding new content, will stabilize

3. **Torrent seeding** (qBittorrent):
   → Limit upload speed in qBittorrent settings

4. **Memory swapping**:
   ```bash
   free -h  # Check available RAM
   ```
   → Increase system RAM or reduce services

**Limit CPU usage** (add to docker-compose.yml):
```yaml
services:
  service_name:
    deploy:
      resources:
        limits:
          cpus: '1.0'  # Limit to 1 CPU core
```

### Database Locked Errors

**Problem**: "Database is locked" errors in logs

**Symptoms**:
```
Error: database is locked
Failed to update: database locked
```

**Solutions**:

1. **Stop the service**:
   ```bash
   docker compose stop sonarr  # or affected service
   ```

2. **Backup the database**:
   ```bash
   cp config/sonarr/*.db config/sonarr/backup/
   ```

3. **Check for corruption**:
   ```bash
   sqlite3 config/sonarr/sonarr.db "PRAGMA integrity_check;"
   ```

4. **Restart service**:
   ```bash
   docker compose start sonarr
   ```

5. **If corruption detected**, restore from backup:
   ```bash
   # Corsarr creates automatic backups
   ls -lah config/sonarr/Backups/
   ```

### Can't Access Web UI

**Problem**: Web interface not loading

**Troubleshooting**:

1. **Check if container is running**:
   ```bash
   docker ps | grep service_name
   corsarr health
   ```

2. **Verify correct port**:
   ```bash
   # Check docker-compose.yml for port mappings
   grep -A 2 "service_name:" docker-compose.yml
   ```

3. **Check firewall**:
   ```bash
   # Allow port in firewall
   sudo ufw allow 7878  # Example for Radarr
   ```

4. **Test from command line**:
   ```bash
   curl http://localhost:7878
   ```

5. **Check container logs**:
   ```bash
   docker compose logs service_name
   ```

### Disk Space Issues

**Problem**: Running out of disk space

**Check usage**:
```bash
df -h /path/to/media
du -sh /path/to/media/*
```

**Solutions**:

1. **Clean up completed downloads**:
   ```bash
   # qBittorrent automatically removes after seeding
   # Or manually delete from downloads folder
   ```

2. **Remove old backups**:
   ```bash
   find config/*/Backups -mtime +30 -delete  # Older than 30 days
   ```

3. **Clean Docker cache**:
   ```bash
   docker system prune -a --volumes
   ```

4. **Monitor with FileFlows** to auto-transcode and compress media

### Network Mode Issues (VPN)

**Problem**: Services not accessible when using VPN mode

**Expected behavior**:
- In VPN mode, ALL traffic goes through Gluetun
- Port mappings must be on Gluetun container, not individual services
- Services communicate via container names

**Verify configuration**:
```bash
# All services should have network_mode: "service:gluetun"
grep -A 1 "network_mode" docker-compose.yml

# Ports should only be on gluetun
grep -B 3 "ports:" docker-compose.yml
```

**If ports aren't working**:
1. Stop everything: `docker compose down`
2. Regenerate with VPN: `corsarr generate --vpn`
3. Verify Gluetun has all necessary ports
4. Start: `docker compose up -d`

### API Key Issues

**Problem**: Services can't authenticate with each other

**Solutions**:

1. **Get API key** from service settings:
   - Go to Settings → General → Security
   - Copy API Key

2. **Add to dependent service**:
   - In Radarr → Settings → Indexers → Add Prowlarr
   - Paste Prowlarr's API key

3. **Verify API key** is correct:
   ```bash
   # Test with curl
   curl -H "X-Api-Key: YOUR_KEY" http://localhost:9696/api/v1/system/status
   ```

### Slow Performance

**Problem**: Services are sluggish or unresponsive

**Optimizations**:

1. **Check system resources**:
   ```bash
   htop  # or top
   free -h
   df -h
   ```

2. **Reduce concurrent operations**:
   - Limit simultaneous downloads in qBittorrent
   - Reduce RSS sync frequency in Sonarr/Radarr
   - Schedule library scans during off-hours

3. **Use SSD for database** (config directory):
   ```yaml
   volumes:
     - /fast/ssd/config:/config  # SSD
     - /slow/hdd/media:/data     # HDD is fine
   ```

4. **Increase container memory** if swapping occurs:
   ```yaml
   deploy:
     resources:
       limits:
         memory: 2G
   ```

### Need More Help?

**Before asking for help**, collect this information:

```bash
# System info
uname -a
docker --version
docker compose version

# Corsarr info
corsarr health --detailed > health-report.txt

# Service logs (last 100 lines)
docker compose logs --tail=100 > service-logs.txt

# Configuration (remove sensitive data!)
cat docker-compose.yml > my-compose.txt
cat .env | grep -v "PRIVATE_KEY\|PASSWORD" > my-env.txt
```

**Where to get help**:
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
