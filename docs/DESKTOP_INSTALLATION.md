# Install and maintain Corsarr Desktop

Corsarr Desktop `1.2.0` supports macOS 14 or newer on Apple Silicon and Intel.
The application manages only resources labeled as owned by Corsarr and keeps
application configuration separate from downloaded media.

## Install on macOS

1. Download
   [`corsarr_desktop_darwin_universal.zip`](https://github.com/woliveiras/corsarr/releases/latest/download/corsarr_desktop_darwin_universal.zip)
   from the [latest GitHub Release](https://github.com/woliveiras/corsarr/releases/latest).
2. Open the archive and move `Corsarr.app` to `Applications`.
3. Because this project does not use an Apple Developer ID, macOS cannot
   notarize the application. On the first launch, Control-click `Corsarr.app`,
   choose **Open**, review the warning, and choose **Open** again. If macOS does
   not offer that action, open **System Settings > Privacy & Security** and use
   **Open Anyway** for Corsarr.
4. Complete the one-time onboarding. Review the Docker and application terms
   before authorizing downloads or installation.

Only download Corsarr from the official `woliveiras/corsarr` GitHub Releases
page and compare the archive with `desktop_checksums.txt` from the same release.
The workflow also publishes an artifact attestation and an SPDX JSON SBOM.

## Update manually

1. Close Corsarr Desktop. Closing the interface does not stop running media
   services.
2. Download the newer archive and verify its checksum.
3. Replace `Corsarr.app` in `Applications` with the newer copy.
4. Open Corsarr. Existing setup state, application configuration, and media are
   outside the application bundle and remain available.

Corsarr updates managed media applications separately from updates to Corsarr
Desktop. Before an application update it backs up configuration, replaces the
approved image, checks health, and restores the previous container image when
possible. A container rollback cannot reverse a database migration made by the
application itself.

## Uninstall

Removing `Corsarr.app` removes only the Desktop interface. It does not silently
delete Docker Desktop, containers, application configuration, downloads, or
media.

For a complete cleanup, first use Corsarr to remove each installed application.
Choose separately whether to archive its configuration. Review and move or
delete the selected Corsarr storage folder yourself only after confirming that
its `media` and `downloads` contents are no longer needed. Docker Desktop must
also be uninstalled separately using Docker's official procedure.

## Data locations and preservation

- Corsarr setup state is stored in the current user's application configuration
  directory.
- The folder selected during onboarding contains application configuration,
  downloads, and media in separate subdirectories.
- Removing a container preserves its configuration by default.
- Archiving application configuration is a separate, explicit action and never
  targets the shared media or downloads tree.
- Exported diagnostic reports are written only to the location selected by the
  user and redact credential-shaped values.

Back up the selected storage folder before operating-system migration, disk
replacement, or major application upgrades.

## Known limitations

- The macOS archive is ad-hoc signed, not Apple Developer ID signed or
  notarized. Gatekeeper therefore requires the explicit first-open action above.
- Automatic Desktop self-update is not implemented; replace the app manually.
- Windows and Linux builds are experimental. The source currently blocks the
  automatic runtime preparation flow and does not provide the required native
  credential store on those platforms, so they are not supported for the
  non-technical onboarding release.
- Docker Desktop installation can request administrator permission. Docker's
  terms and subscription conditions remain the user's decision.
- Disabling virtualization, resolving VPN/proxy/CA problems, and granting macOS
  access to protected folders may require operating-system action.
- Hardlinks require downloads and libraries to remain on the same compatible
  filesystem. Corsarr reports missing hardlink support as an efficiency warning.
- Removing an older container image does not undo an application database
  migration.

For errors, see [Troubleshooting](TROUBLESHOOTING.md) and attach the exported
technical report to a [bug report](https://github.com/woliveiras/corsarr/issues/new/choose).
