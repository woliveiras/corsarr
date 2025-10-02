# Stack*RR

> Fast way for running for Radarr, Lidarr, LazyLibrarian, and Bazarr

This repository contains a Docker Compose setup for managing various media applications, including Radarr, Lidarr, LazyLibrarian, and Bazarr. The setup is designed to run on a Linux server with Docker and Docker Compose installed.

In this guide, you’ll learn how to set up a fully automated media server that, when you add a movie or TV show, will:

- Search for torrents/magnet links using Prowlarr and Jellyseerr (optional)
- Manage content with Sonarr (TV shows), Radarr (movies), LazyLibrarian (books), and Lidarr (music)
- Download subtitles with Bazarr
- Download files via qBittorrent
- FlareSolverr to bypass Cloudflare restrictions
- Stream your media to your TV with Jellyfin

All you need is my `docker-compose.yml` file and a few environment variables!

## Prerequisites

- [Docker and Docker Compose](https://docs.docker.com/compose/install/#docker-desktop-recommended) installed on your server;

## Setup Instructions

1. **Clone the Repository**: Clone this repository to your server using the following command:

   ```bash
   mkdir servarr && cd servarr

   git clone git@github.com:woliveiras/stack-rr.git .
   ```

   If you prefer using HTTPS, use this command instead:

   ```bash
   mkdir servarr && cd servarr

   git clone https://github.com/woliveiras/stack-rr.git .
   ```

1. **Navigate to the Directory**: we can use two settings here:

   - `simple`: without VPN

      ```bash
      cd simple
      ```

   - `vpn`: with VPN (you need to configure your VPN settings in the `.env` file)

      ```bash
      cd vpn
      ```

1. **Configure Environment Variables**: You need to set up your environment variables. Change the name of the file `.env.example` to `.env`. Set the following variables:

   ```env
   ARRPATH=/arr/      # Path to your media library, keep the final slash
                      # Example: ARRPATH=/path/to/your/media/ 
                      # /Users/<User>/Downloads/RADARR/
   TZ=Europe/Madrid   # Set your timezone https://www.timeanddate.com/time/map/

   # You don't need to change the following ones unless you have a specific reason
   PUID=1000
   PGID=1000
   UMASK=002
   ```

   If you are using the VPN setup, you also need to configure your VPN settings in the `.env` file. Check to the comments in the file for guidance.


## Running the Stack

1. **Start the Stack**: Use Docker Compose to start the stack. Run the following command in the directory where your `docker-compose.yml` file is located:

   ```bash
   docker-compose up -d
   ```

   This command will pull the necessary images and start the containers in detached mode.

   The first time you run this command, it may take a while to download all the images and set up the containers.

2. **Access the Applications**: Once the containers are running, you can access the applications via your web browser:

   - Radarr: `http://<your-server-ip>:7878`
   - Lidarr: `http://<your-server-ip>:8686`
   - LazyLibrarian: `http://<your-server-ip>:5299`
   - Bazarr: `http://<your-server-ip>:6767`
   - qBittorrent: `http://<your-server-ip>:8081`
   - Prowlarr: `http://<your-server-ip>:9696`
   - Jellyfin: `http://<your-server-ip>:8096`
   - Jellyseerr (optional): `http://<your-server-ip>:5055`

   Ex.: Radarr: `http://localhost:7878/` for local access or `192.168.1.x:7878/` for network access (your computer need to be available in the network for this option).

3. **Stopping the Stack**: To stop the stack, run:

   ```bash
   docker-compose down
   ```

## Initial Configuration of the *arr Suite

After accessing each application for the first time, you will need to go through the initial setup process. This typically involves setting up your media library paths, configuring download clients, and adding indexers.

###  qBittorrent

Access `http://<YOUR_SERVER>:8081` (default user: `admin`, password: run `docker logs qbittorrent` in your terminal).

You can change the credentials in **Settings** > **Web UI** > **Authentication** and click Save.

Set the download path to `/downloads` in Tools > Options > Downloads and click Save.

You don't need to paste magnet links in qBittorrent. The *arr applications will do that for you.

### Radarr, Lidarr, LazyLibrarian, Sonarr

Access `http://<YOUR_SERVER>:7878` (Radarr) or `8989` (Sonarr).

Go to **Settings** > **Media Management**, click on **Add Root Folder** and add /data/movies for Radarr or /data/tvshows for Sonarr.

Go to **Download Clients** and click on the plus button (+): add *qBittorrent* (host: `qbittorrent`, port: `8081`, your username and password).

Copy the API Key from **General** > **Security** > API Key.

### Prowlarr

Access `http://<YOUR_SERVER>:9696`

Go to **Settings** > **Download Clients** and click on the plus button (+): add *qBittorrent* (host: `qbittorrent`, port: `8081`, your username and password).

Go to **Settings** > **Apps**: add Radarr, Lidarr, LazyLibrarian and Sonarr using their URLs and API Keys you copied earlier. In the Prowlarr Server, use `http://Prowlarr:9696`, and in the Radarr/Lidarr/LazyLibrarian/Sonarr Server, use `http://Lidarr:8686`, `http://Radarr:7878`, `http://LazyLibrarian:5299` or `http://Sonarr:8989`.

### Bazarr

Access `http://<YOUR_SERVER>:6767`

Go to **Settings** > **Providers**: add the subtitle providers you want to use.

Go to **Settings** > **Radarr** and add Radarr (address: `radarr`, port: `7878`, base URL: `/`, your API Key).

Go to **Settings** > **Sonarr** and add Radarr (address: `sonarr`, port: `8989`, base URL: `/`, your API Key).

For Multi-language Search: In Settings > Profiles, edit profiles and change Language to `<Language You Want>` to search for `<Language>` content.

### FlareSolver

You'll need to configure FlareSolverr in each *arr application to bypass Cloudflare restrictions.

Go to **Settings** > **Indexers** and edit each indexer you have added (you may need to add them first if you haven't done so).

### Jellyfin

Access `http://<YOUR_SERVER>:8096`

Go through the initial setup process, creating an admin user and adding your media libraries.

You can install the Jellyfin app on your smart TV, phone, or tablet to stream your media. When asked for the server address, use `http://<YOUR_SERVER>:8096`. The username and password are the ones you created during the initial setup.

### Jellyseerr (Optional)

Jellyseerr is a request management system for movies and TV shows that integrates with Radarr, Sonarr, and Plex/Jellyfin.

It allows users to request new content, view existing requests, and manage their media library through a user-friendly interface.

Access `http://localhost:5055/` and follow the initial setup instructions.

## Demo: Downloading Public Domain Content

In Radarr, try downloading public domain movies:

Click on **Add New Movie** and search for `Night of the Living Dead (1968)`.

Watch as it changes to Downloading and appears in qBittorrent.

Repeat with `Nosferatu (1922)` and `The City of the Dead (1960)`.

In Jellyfin/Plex, you’ll see the movies with covers, synopsis, and actors ready to play.

You can find more public domain movies at [JustWatch](https://www.justwatch.com/us/provider/public-domain-movies?sort_by=trending_7_day).

For public domain books, you can use [goodreads](https://www.goodreads.com/shelf/show/public-domain).

## Update Instructions

To update the stack, navigate to the directory where your `docker-compose.yml` file is located and run:

```bash
docker-compose pull
docker-compose up -d
```

This will pull the latest images and restart the containers with the updated versions.

---

## Example .env File

Here is a sample `.env` file for quick reference:

```env
ARRPATH=/Users/<User>/Downloads/RADARR/data/   # Path to your media library, keep the final slash
TZ=Europe/Madrid                                # Set your timezone
PUID=1000                                       # User ID for permissions
PGID=1000                                       # Group ID for permissions
UMASK=002                                       # File creation mask
```

---

## Backup and Restore Instructions

To backup your configuration and databases, simply copy the relevant folders from the `config/` directory (e.g., `config/radarr/`, `config/sonarr/`, etc.) and the `data/` directory for your media files.

**Backup Example:**

```bash
cp -r config/radarr/ ~/radarr-backup/
cp -r data/movies/ ~/movies-backup/
```

**Restore Example:**

```bash
cp -r ~/radarr-backup/ config/radarr/
cp -r ~/movies-backup/ data/movies/
```

For automated backups, consider using tools like [Duplicati](https://www.duplicati.com/) or [rsync](https://rsync.samba.org/).

---

## Official Documentation Links

- [Radarr Documentation](https://wiki.servarr.com/radarr)
- [Sonarr Documentation](https://wiki.servarr.com/sonarr)
- [Lidarr Documentation](https://wiki.servarr.com/lidarr)
- [LazyLibrarian Documentation](https://lazylibrarian.gitlab.io/)
- [Bazarr Documentation](https://wiki.bazarr.media/)
- [qBittorrent Documentation](https://github.com/qbittorrent/qBittorrent/wiki)
- [Jellyfin Documentation](https://jellyfin.org/docs/)
- [Jellyseerr Documentation](https://docs.jellyseerr.com/)

---

## Useful Docker Commands

Check running containers:

```bash
docker-compose ps
```

View logs for a specific container:

```bash
docker logs <container_name>
```

Restart a container:

```bash
docker-compose restart <service_name>
```

Stop all containers:

```bash
docker-compose down
```

---

## Security Recommendations

- **Restrict External Access:** Only expose necessary ports. Use firewall rules to block unwanted traffic.
- **Use VPN:** For remote access, prefer VPN solutions (see the VPN Setup section above).
- **Update Regularly:** Keep your Docker images up to date (`docker-compose pull`).
- **Change Default Passwords:** Always change default credentials for all services.
- **Use HTTPS:** If exposing services externally, consider using a reverse proxy (e.g., Nginx, Traefik) with SSL certificates.

---

## Customizing Volumes and Paths

You can change the default paths for your media and configuration by editing the `docker-compose.yml` and `.env` files. For example, to use a different folder for movies, update the volume mapping:

```yaml
   radarr:
      ...existing code...
      volumes:
         - /custom/path/to/movies:/data/movies
         - /custom/path/to/config:/config/radarr
```

Make sure the user running Docker has read/write permissions to these folders. Adjust `PUID`, `PGID`, and `UMASK` in your `.env` file as needed.

---

## FAQ: Common Problems and Solutions

### 1. Permission Issues

- *Symptoms:* Containers fail to access files or folders, errors about permissions.
- *Solution:* Ensure your media/config folders are owned by the correct user/group. Use `chown` and `chmod` as needed:

   ```bash
   sudo chown -R 1000:1000 /path/to/media /path/to/config
   sudo chmod -R 775 /path/to/media /path/to/config
   ```

### 2. Ports Already in Use

- *Symptoms:* Docker fails to start containers due to port conflicts.
- *Solution:* Change the exposed ports in `docker-compose.yml` or stop the conflicting service.

### 3. Containers Not Starting

- *Symptoms:* `docker-compose up` fails, containers exit immediately.
- *Solution:* Run `docker-compose logs <service_name>` to check for errors. Common issues include missing environment variables, permission problems, or port conflicts.

### 4. Slow Downloads or No Connection

- *Symptoms:* qBittorrent or *arr apps can't download or connect to indexers.
- *Solution:* Check your network settings, VPN configuration, and firewall rules. Make sure your server has internet access.

### 5. Lost Configuration or Data

- *Symptoms:* Settings or media disappear after restart.
- *Solution:* Ensure your volumes are mapped correctly in `docker-compose.yml` and data is stored outside the container.

### 6. Web UI Not Accessible

- *Symptoms:* Can't access service via browser.
- *Solution:* Confirm the container is running (`docker-compose ps`), check the correct port, and verify firewall settings.

For more troubleshooting, consult the official documentation links above.
