# Run Corsarr without Docker Desktop

Corsarr Desktop and CLI can share a local Docker Engine. Docker Desktop remains
the default for existing installations. No migration of existing containers or
libraries is performed when configuring a new environment.

## Existing local Docker

Start your approved Docker environment and list its contexts:

```sh
docker context ls
corsarr runtime use existing --context default
corsarr runtime status
```

Replace `default` with the local context name, for example `rootless` or `colima`.
The Docker CLI must be on Corsarr's PATH. The daemon must run Linux containers.
The context must resolve to a local Unix socket or local Windows named pipe;
SSH and TCP endpoints are rejected because this release uses local storage and
local application URLs. A VM must expose the selected storage folders to Docker;
detecting its socket does not prove filesystem sharing works.

Corsarr does not install, start or recover an externally managed runtime. Start
that environment yourself or through your administrator's approved tooling.
Corsarr does not change Docker's global current context.

## Install Docker Engine on Linux

Automatic installation supports Ubuntu 22.04, 24.04 and 26.04, and Debian 12 and
13, on amd64/arm64 with systemd running. Derivative distributions are not included.
Other Linux systems can use an existing installation instead.

In a new Desktop setup, choose **Install Docker Engine on Linux** in the permissions
step, save the choice and accept the authorization. Preparing the environment
requests administrator authorization through PolicyKit (`pkexec` and a graphical
authentication agent must be installed).

For the CLI, review the effects below and then run:

```sh
sudo corsarr runtime install --yes

# Run these as your regular user, not under sudo:
corsarr runtime use docker-engine
corsarr runtime status
```

Alternatively, after selecting `docker-engine`, `corsarr runtime prepare --yes`
uses PolicyKit to install Engine or start an existing system Docker service.
`runtime install` installs packages only; it does not save root's configuration
as the configuration for your regular account.

The installer:

- Adds the official signed stable Docker APT repository for the detected release.
- Verifies the repository key against Docker's expected key fingerprint.
- Installs Docker Engine, CLI, containerd, Buildx and the Compose plugin.
- Enables and starts the system Docker service.
- Refuses existing Docker installations and conflicting packages without removing them.
- Leaves account groups, existing containers and application data untouched.

Package versions are the candidates provided by the configured APT repositories;
they are not a Corsarr-pinned runtime release. Interrupted package operations may
require administrator repair before retrying. Installation errors are returned
to the user; no destructive rollback or package purge is attempted.

If installation succeeds but Docker reports **permission denied**, ask your
administrator to configure socket access. Adding a user to the `docker` group
grants root-equivalent privileges and is not done automatically. Sign out and
back in after an administrator changes group membership. A rootless Docker setup
can instead be selected through its local context using `existing`.

## Use the same environment from the CLI

Generation still produces Compose files. To run them against Corsarr's selected
environment, use the wrapper instead of an unqualified `docker compose` command:

```sh
corsarr runtime compose -f ./output/docker-compose.yml up -d
corsarr runtime compose -f ./output/docker-compose.yml ps
corsarr runtime compose -f ./output/docker-compose.yml logs --tail 100
```

The wrapper buffers output until the command completes (up to 30 minutes);
use bounded log queries rather than `logs --follow`. `corsarr health` and Docker
validation also use the selection. Compose v2 or newer is required.

Configuration is stored in `Corsarr/execution.json` under the user's OS
configuration directory. Close Desktop before changing it through the CLI and
reopen Desktop afterward. The destination is locked once Desktop storage or
applications have been selected; switching an existing installation requires a
future migration workflow. Changing the environment clears prior runtime consent.

## Scope and verification

This delivery includes existing local Docker and native Linux installation.
It does not install WSL, Colima, Hyper-V or Podman and does not support remote
servers. Those remain separate follow-up work.

Tests cover selection persistence, endpoint isolation, remote endpoint rejection,
Windows-container rejection, consent and storage locks, permission failures and
installer control flow using isolated fake system commands. Real fresh-host
Ubuntu/Debian installation and desktop PolicyKit authorization still require
release validation on those systems.

Official installation references:
[Ubuntu](https://docs.docker.com/engine/install/ubuntu/),
[Debian](https://docs.docker.com/engine/install/debian/), and
[post-installation permissions](https://docs.docker.com/engine/install/linux-postinstall/).
