# rexctl

A CLI tool for managing multi-repository Docker Compose workspaces.

`rexctl` coordinates workspaces that consist of multiple Git repositories (with Docker Compose setups) and standalone container images on a single machine. It automates cloning, revision checkouts, standard `docker-compose.override.yml` generation with commit tracking, environment initialization, and container lifecycle management.

## Installation

### Nix / NixOS

`rexctl` is packaged as a Nix flake and includes a NixOS module:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    rexctl.url = "github:BlacksQare/rexctl";
  };

  outputs = { self, nixpkgs, rexctl }: {
    nixosConfigurations.my-host = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        rexctl.nixosModules.default
        {
          programs.rexctl = {
            enable = true;
            workspacesPath = "/var/lib/rexctl/workspaces";
            defaultShellUser = "rex";
          };
        }
      ];
    };
  };
}
```

### Manual Build

Requires Go 1.21+:

```bash
go build -o rexctl .
```

## Quick Start

1. Create a workspace:
   ```bash
   rexctl create my-workspace
   ```
2. Edit the manifest:
   ```bash
   rexctl edit my-workspace
   ```
3. Clone repositories and prepare compose overrides:
   ```bash
   rexctl sync my-workspace
   ```
4. Build and start containers:
   ```bash
   rexctl up my-workspace
   ```
5. Connect to a container shell:
   ```bash
   rexctl sh ros-core
   ```

## Workspace Manifest (`rex.yaml`)

Workspaces are defined in a `rex.yaml` file located in each workspace directory:

```yaml
kind: RexctlWorkspace

spec:
  containers:
    - name: example_voting_app
      type: compose
      remote: https://github.com/dockersamples/example-voting-app
      revision: main

    - name: nginx
      type: image
      remote: nginx:stable-trixie-perl
```

### Container Types

- `compose`: Git repository containing Docker Compose files. `rexctl` clones the repository, checks out the specified `revision`, generates `docker-compose.override.yml` with commit/dirty labels, and runs Docker Compose in that directory.
- `image`: Standalone container image pulled directly from a container registry and managed under the workspace project name.

## Commands

| Command | Usage | Description |
| :--- | :--- | :--- |
| `create` | `rexctl create <workspace>` | Initialize a new workspace directory with a default manifest |
| `edit` | `rexctl edit [workspace]` | Open `rex.yaml` in `$EDITOR` |
| `sync` | `rexctl sync [workspace]` | Fetch/clone repositories, checkout revisions, pull images, and generate overrides |
| `prepare-env` | `rexctl prepare-env [workspace]` | Run `rexctl_init.sh` across all cloned repositories |
| `build` | `rexctl build [workspace]` | Build container images without starting them (`docker compose build`) |
| `up` | `rexctl up <workspace>` | Build and start containers (`docker compose up -d --build`) |
| `start` | `rexctl start <workspace>` | Start existing stopped containers (`docker compose start`) |
| `stop` | `rexctl stop [workspace]` | Gracefully stop running containers without removing them |
| `down` | `rexctl down [workspace]` | Stop and remove containers and networks (`docker compose down`) |
| `switch` | `rexctl switch <workspace>` | Stop current workspace and start the target workspace (without rebuilding) |
| `destroy` | `rexctl destroy <workspace>` | Remove containers and delete the workspace directory |
| `shell` / `sh` | `rexctl sh [-u user] [workspace] <container>` | Open an interactive shell in a container |
| `get` | `rexctl get` | Print the active workspace name (or empty if inactive) |
| `status` | `rexctl status` | Print running container count for the active workspace |
| `info` | `rexctl info [workspace]` | Print workspace details, revisions, git status, and image tags |
| `list` | `rexctl list` | List all available workspaces |
| `pwd` | `rexctl pwd [workspace]` | Print the filesystem path of a workspace |
| `validate` | `rexctl validate [workspace]` | Validate `rex.yaml` syntax and structure |

If `[workspace]` is omitted on commands that support it, `rexctl` infers the workspace from the current working directory (if inside a workspace) or the currently active stack.

## Lifecycle Concepts

### `build` vs `up` vs `start`
- `rexctl build [workspace]`: Builds images (`docker compose build`) and refreshes `docker-compose.override.yml` without starting containers.
- `rexctl up <workspace>`: Builds images, creates container instances, and starts them (`docker compose up -d --build`). Refreshes `docker-compose.override.yml` with the latest commit hash and dirty status before starting.
- `rexctl start <workspace>`: Resumes existing, stopped containers (`docker compose start`) without rebuilding.

### `stop` vs `down`
- `rexctl stop [workspace]`: Pauses containers while preserving container state and internal volumes (`docker compose stop`).
- `rexctl down [workspace]`: Stops and removes containers and networks (`docker compose down`).

### Compose Overrides
During `sync`, `build`, and `up`, `rexctl` inspects compose files and generates a `docker-compose.override.yml` in each repository:
- Buildable services are tagged with `rexctl/<workspace>/<service>:<commit>[-dirty]`.
- Pre-built services retain their image names while attaching metadata labels (`rexctl.workspace`, `rexctl.service`, `rexctl.commit`, `rexctl.dirty`).

## Development

```bash
# Run tests
go test -v ./...

# Build
go build -o rexctl .
```