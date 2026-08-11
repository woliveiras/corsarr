# Operate a CLI-generated media stack

This guide applies to stacks generated with the Corsarr CLI. Corsarr Desktop
automates many of these connections through its own setup flow.

Use only applications, indexers, downloads, and media that you are legally
authorized to access.

## Start and stop

Run these commands from the directory containing `docker-compose.yml`:

```bash
docker compose up -d
docker compose ps
docker compose down
```

`docker compose down` stops and removes containers and networks. It does not
delete bind-mounted application configuration or media.

## Local application addresses

The default generated ports are:

| Application | Address |
| --- | --- |
| qBittorrent | `http://localhost:8081` |
| Prowlarr | `http://localhost:9696` |
| Radarr | `http://localhost:7878` |
| Sonarr | `http://localhost:8989` |
| Bazarr | `http://localhost:6767` |
| Jellyfin | `http://localhost:8096` |
| Seerr | `http://localhost:5055` |

Ports can differ when a conflict was resolved during generation. Check the
generated Compose file or run `corsarr check-ports`.

## Initial application setup

For a CLI-generated stack, configure the selected applications through their
local web interfaces:

1. Set qBittorrent's download path and replace its temporary administrator
   password.
2. Add permitted indexers and qBittorrent to Prowlarr.
3. Add the generated movie and TV root folders to Radarr and Sonarr, then add
   qBittorrent and Prowlarr using their local service names and credentials.
4. Connect Bazarr to Radarr and Sonarr when subtitles are selected.
5. Create the Jellyfin administrator and add only the generated media folders
   that should be available in the library.
6. Connect Seerr to Jellyfin, Radarr, and Sonarr when requests are selected.

Never expose administrative web interfaces directly to the public internet.
Use platform documentation for the exact fields because application screens
and authentication requirements change independently of Corsarr.

## Update applications

Back up configuration before updating, then run:

```bash
docker compose pull
docker compose up -d
docker compose ps
```

Review application release notes before major upgrades. Replacing a container
image cannot reverse a database migration performed by the application.

## Back up and restore

Stop writes or stop the stack before copying live databases. Back up the
generated configuration directories and media to storage outside the stack.
The exact paths are the ones selected during `corsarr generate`.

After restoring the same directory structure and generated files:

```bash
docker compose up -d
docker compose ps
```

Application-provided backups may also exist below their own configuration
directories, but they do not replace a complete external backup.

## Security

- Replace temporary or default credentials.
- Bind administrative interfaces to trusted networks only.
- Keep Corsarr, container images, the operating system, and the container
  runtime updated.
- Protect `.env`, profiles, VPN credentials, API keys, and backups.
- Use HTTPS through a reviewed reverse proxy before enabling remote access.
- Review firewall, VPN, proxy, and DNS behavior for the host network.

For operational failures, see [Troubleshooting](TROUBLESHOOTING.md).
