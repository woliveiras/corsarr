# 🆘 Troubleshooting Guide

Use this guide to diagnose common issues when running Corsarr-generated stacks.

## Service Can't Access Files

**Problem**: Permission denied errors

```bash
# Fix ownership
sudo chown -R 1000:1000 /path/to/media

# Verify PUID/PGID match
id $(whoami)
```

## Port Already in Use

**Problem**: Port conflict errors (e.g., "address already in use")

```bash
# Check which ports are in use
corsarr check-ports --suggest

# Check specific port
sudo lsof -i :8080

# Or check all Docker ports
docker ps --format "table {{.Names}}\t{{.Ports}}"
```

**Common conflicts**:
- **Port 1900** - DLNA/UPnP service (minidlna, Jellyfin host)
  - Solution: Stop the conflicting service or remove from docker-compose.yml
- **Port 8080/8081** - Web services (qBittorrent, other apps)
  - Solution: Change port in docker-compose.yml

## VPN Not Working

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
- **Verify NET_ADMIN capability**: Ensure `cap_add: - NET_ADMIN` is present under the Gluetun service

## Container Won't Start

**Problem**: Service keeps restarting or "EOF" / "can't get final child's PID" errors

**Common causes**:
1. **Missing NET_ADMIN for Gluetun**: VPN won't work without this capability.
2. **Volume permission issues**: Run `sudo chown -R $USER:$USER /path/to/media`.
3. **Service dependency failed**: Check if a dependent container (e.g., Gluetun) is healthy.

```bash
# Check health status
corsarr health --detailed

# View service logs
docker compose logs [service_name]

# Check Gluetun specifically (if using VPN)
docker compose logs gluetun | grep -i error

# Check for errors
docker compose ps
```

## Can't Connect to Other Services

**Problem**: Radarr can't reach Prowlarr

**Solution**: Use container names, not localhost.
- ✅ `http://prowlarr:9696`
- ❌ `http://localhost:9696`

## High CPU Usage

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

## Database Locked

**Problem**: "Database is locked" errors

```bash
# Stop the affected service
docker compose stop sonarr

# Backup database
cp config/sonarr/*.db config/sonarr/backup/

# Restart service
docker compose start sonarr
```

## Need More Help?

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
