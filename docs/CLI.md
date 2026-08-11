# Install and use the Corsarr CLI

The Corsarr CLI generates `docker-compose.yml`, `.env`, and the directories
needed by a selected media stack. Docker with Docker Compose v2 is required to
run the generated stack.

## Install

Download the archive for your operating system from the
[latest release](https://github.com/woliveiras/corsarr/releases/latest).

### Linux x64

```bash
release=https://github.com/woliveiras/corsarr/releases/latest/download
curl -sL "$release/corsarr_linux_amd64.tar.gz" | tar xz
sudo mv corsarr /usr/local/bin/corsarr
```

Use `corsarr_linux_arm64.tar.gz` instead on Linux ARM64.

### macOS

Apple Silicon:

```bash
release=https://github.com/woliveiras/corsarr/releases/latest/download
curl -sL "$release/corsarr_darwin_arm64.tar.gz" | tar xz
sudo mv corsarr /usr/local/bin/corsarr
```

Intel:

```bash
release=https://github.com/woliveiras/corsarr/releases/latest/download
curl -sL "$release/corsarr_darwin_amd64.tar.gz" | tar xz
sudo mv corsarr /usr/local/bin/corsarr
```

### Windows x64

In an elevated PowerShell session:

```powershell
$release = "https://github.com/woliveiras/corsarr/releases/latest/download"
$archive = "$env:TEMP\corsarr.zip"
Invoke-WebRequest -Uri "$release/corsarr_windows_amd64.zip" -OutFile $archive
Expand-Archive -Path $archive -DestinationPath "C:\Program Files\corsarr" -Force
```

Add `C:\Program Files\corsarr` to the system `PATH`, then open a new terminal.
The release also provides `corsarr_windows_386.zip` for x86 systems.

Confirm the installation:

```bash
corsarr --version
```

## Generate and start a stack

Run the interactive setup in an empty directory:

```bash
corsarr generate
docker compose up -d
```

The setup asks for language, applications, storage, timezone, user and group
IDs, and optional VPN details. Use only services and sources that you are
legally authorized to access.

To generate files in another directory:

```bash
corsarr generate --output ~/my-media-stack
```

Preview without writing files:

```bash
corsarr generate --dry-run
```

Select a language explicitly:

```bash
corsarr --language pt-br generate
corsarr --language es generate
```

## Inspect a generated stack

```bash
corsarr health
corsarr health --detailed
corsarr check-ports
corsarr check-ports --suggest
corsarr preview
```

Pass `--output /path/to/stack` to `health` or `check-ports` when the Compose
files are not in the current directory.

## Profiles

Save the result of an interactive generation:

```bash
corsarr generate --save-profile --save-as my-setup
```

Reuse or manage profiles:

```bash
corsarr generate --profile my-setup
corsarr profile list
corsarr profile load my-setup
corsarr profile export my-setup backup.json
corsarr profile import backup.json --name restored-setup
corsarr profile delete my-setup
```

## Non-interactive generation

Automation must provide the complete configuration explicitly:

```bash
corsarr generate --no-interactive \
  --services "prowlarr,radarr,sonarr,jellyfin,qbittorrent" \
  --arr-path "/home/user/media" \
  --timezone "Europe/Madrid" \
  --puid "1000" \
  --pgid "1000" \
  --output ./stack
```

Configuration can also come from YAML or JSON:

```yaml
services:
  - prowlarr
  - radarr
  - sonarr
  - jellyfin
  - qbittorrent
arr_path: /home/user/media
timezone: Europe/Madrid
puid: 1000
pgid: 1000
```

```bash
corsarr generate --config config.yaml --no-interactive
```

Run `corsarr generate --help` for the authoritative list of configuration,
VPN, profile, and automation flags.

## Update

Download the latest archive for the same platform and replace the installed
executable. This does not alter previously generated files.

Updating Corsarr is separate from updating the applications in a generated
stack. See [Stack operations](STACK_OPERATIONS.md).

## Uninstall

Remove only the installed executable:

- Linux or macOS: remove `/usr/local/bin/corsarr`, or the exact path chosen
  during installation.
- Windows: remove `C:\Program Files\corsarr` and remove that directory from
  `PATH`.

Uninstalling the CLI does not stop containers or delete generated files,
configuration, downloads, or media. Run `docker compose down` in the generated
stack directory first if the containers should be stopped.
