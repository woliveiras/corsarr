# Corsarr 🏴

> 🏴‍☠️ Navigate the high seas of media automation

<p align="center">
  <img src="assets/corsarr-logo-transparent.png" alt="Corsarr Logo" width="300"/>
</p>

Corsarr helps you install and operate the applications in a local media server.

Use the visual Desktop application or the CLI.

## Corsarr Desktop

### Download Desktop

| Platform | Download |
| --- | --- |
| macOS — Apple Silicon and Intel | [Universal ZIP](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_desktop_darwin_universal.zip) |
| Windows — x64 | [Portable ZIP](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_desktop_windows_amd64.zip) |
| Windows — x86 | [Portable ZIP](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_desktop_windows_386.zip) |
| Linux — x64 | [Portable tar.gz](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_desktop_linux_amd64.tar.gz) |

All downloads, checksums, and release notes are available on the
[latest release page](https://github.com/woliveiras/corsarr/releases/latest).

### Install and use

1. Download and extract the package for your platform.
2. On macOS, move `Corsarr.app` to `Applications`. On Windows or Linux, open
   the extracted Corsarr executable.
3. Follow the setup screens to choose storage and applications, review the
   configuration, and start the installation.
4. Use the dashboard to open and manage installed applications.

macOS may require an explicit first-open action. See the
[Desktop installation guide](docs/DESKTOP_INSTALLATION.md) for platform notes,
checksums, data locations, and complete maintenance instructions.

### Update

Close Corsarr, download the new package, and replace the previous application
or executable. Your media and application data are stored separately.

### Uninstall

Remove `Corsarr.app` or the extracted Corsarr executable. This removes the
interface only; it does not silently delete installed applications, their
configuration, downloads, media, or the container runtime. Use Corsarr to
remove managed applications before deleting data you no longer need.

## Corsarr CLI

### Download CLI

| Platform | Download |
| --- | --- |
| macOS — Apple Silicon | [tar.gz](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_darwin_arm64.tar.gz) |
| macOS — Intel | [tar.gz](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_darwin_amd64.tar.gz) |
| Linux — x64 | [tar.gz](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_linux_amd64.tar.gz) |
| Linux — ARM64 | [tar.gz](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_linux_arm64.tar.gz) |
| Windows — x64 | [ZIP](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_windows_amd64.zip) |
| Windows — x86 | [ZIP](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_windows_386.zip) |

Extract the archive and place `corsarr` or `corsarr.exe` in a directory listed
in your `PATH`. Docker with Docker Compose v2 is required to run the generated
stack.

### Use

```bash
corsarr generate
docker compose up -d
```

Run `corsarr --help` to see the available commands. The
[CLI guide](docs/CLI.md) covers installation commands, profiles, automation,
updates, and removal.

### Update or uninstall

To update, download the newest archive and replace the executable. To uninstall,
remove the executable from the directory where you installed it. Generated
Compose files and media data remain untouched.

## Documentation

- [Documentation index](docs/README.md)
- [Desktop installation and maintenance](docs/DESKTOP_INSTALLATION.md)
- [CLI guide](docs/CLI.md)
- [Operating a CLI-generated media stack](docs/STACK_OPERATIONS.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [Development and validation](docs/DEVELOPMENT.md)

## Support and license

- [Report a problem](https://github.com/woliveiras/corsarr/issues)
- [Report a security vulnerability](SECURITY.md)
- [MIT License](LICENSE)
- [Third-party notices](THIRD_PARTY_NOTICES.md)
