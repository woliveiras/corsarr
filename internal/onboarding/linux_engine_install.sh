#!/bin/sh
# Embedded, fixed installer. No downloaded script is executed.
set -eu
export DEBIAN_FRONTEND=noninteractive
export PATH=/usr/sbin:/usr/bin:/sbin:/bin
[ "$(id -u)" = 0 ] || { echo 'Administrator permission is required.' >&2; exit 1; }
[ -d /run/systemd/system ] || { echo 'A running systemd system is required.' >&2; exit 1; }
. /etc/os-release
case "$ID:$VERSION_CODENAME" in
  ubuntu:jammy|ubuntu:noble|ubuntu:resolute|debian:bookworm|debian:trixie) ;;
  *) echo 'Supported systems: Ubuntu 22.04/24.04/26.04 and Debian 12/13.' >&2; exit 1 ;;
esac
# Never replace another installation or remove conflicting packages.
for package in docker-ce docker.io docker-compose docker-compose-v2 docker-doc docker-buildx podman-docker containerd runc; do
  if dpkg-query -W -f='${Status}' "$package" 2>/dev/null | grep -qx 'install ok installed'; then
    echo "Existing package $package requires administrator review; no packages were removed." >&2
    exit 1
  fi
done
if command -v docker >/dev/null 2>&1 || command -v dockerd >/dev/null 2>&1 || [ -x /usr/local/bin/docker ] || [ -x /snap/bin/docker ]; then
  echo 'An existing Docker installation was found; use the existing runtime option.' >&2
  exit 1
fi
arch=$(dpkg --print-architecture)
case "$arch" in amd64|arm64) ;; *) echo 'Unsupported architecture.' >&2; exit 1 ;; esac
source_file=/etc/apt/sources.list.d/corsarr-docker.sources
key_file=/etc/apt/keyrings/corsarr-docker.asc
# A second repository with a different Signed-By option can break all APT
# operations. Leave administrator-managed Docker sources untouched.
for repository in /etc/apt/sources.list /etc/apt/sources.list.d/*.list /etc/apt/sources.list.d/*.sources; do
  if [ -f "$repository" ] && [ "$repository" != "$source_file" ] && grep -q 'download\.docker\.com' "$repository"; then
    echo 'A Docker APT repository is already configured; ask your administrator to install Engine from that repository.' >&2
    exit 1
  fi
done
source_text="Types: deb
URIs: https://download.docker.com/linux/$ID
Suites: $VERSION_CODENAME
Components: stable
Architectures: $arch
Signed-By: $key_file"
if [ -e "$source_file" ] && [ "$(cat "$source_file")" != "$source_text" ]; then
  echo 'An existing Corsarr repository file differs; ask your administrator to review it.' >&2
  exit 1
fi
apt-get update
apt-get install -y --no-remove ca-certificates curl gnupg
install -m 0755 -d /etc/apt/keyrings
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT HUP INT TERM
curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' --connect-timeout 30 --max-time 120 "https://download.docker.com/linux/$ID/gpg" -o "$work/docker.asc"
fingerprint=$(gpg --batch --homedir "$work" --show-keys --with-colons "$work/docker.asc" | awk -F: '$1 == "fpr" {print $10; exit}')
[ "$fingerprint" = 9DC858229FC7DD38854AE2D88D81803C0EBFCD88 ] || { echo 'Docker repository key fingerprint mismatch.' >&2; exit 1; }
install -m 0644 "$work/docker.asc" "$key_file"
printf '%s\n' "$source_text" > "$source_file"
chmod 0644 "$source_file"
apt-get update
apt-get install -y --no-remove docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
systemctl enable --now docker
