# rexctl

**Declarative Workspace Orchestrator**

`rexctl` is a CLI tool built in Go that declaratively manages complex, multi-repository Docker development environments. It allows you to spin up, synchronize, and switch between completely isolated development workspaces (like ROS arrays, microservices, or robotics stacks) on a single machine without container naming conflicts.

## Features

* **Declarative Environments:** Define your entire workspace (Git repositories, Docker Compose files, and raw images) in a single `rex.yaml` manifest.
* **True Isolation:** Automatically resolves global Docker `container_name` collisions. `rexctl` dynamically injects override files to namespace your containers, allowing you to run multiple instances of the same repository simultaneously.
* **Stateful Workspaces:** Switch between workspaces seamlessly. `rexctl` stops (rather than destroys) containers, preserving your volumes and run states.
* **Auto-Healing & Sync:** Automatically detects corrupted directories, missing compose files, or untracked Git changes, and interactively offers to force-clean and rebuild the state.
* **NixOS First:** Natively supports Nix flakes and NixOS modules. Workspaces directories and default manifests can be baked directly into the binary at compile time via the Nix store.

---

## Quick Start

### 1. Initialize a Workspace
Create a new workspace directory and generate the default manifest:
```bash
rexctl create my-workspace
```

### 2. Edit the Manifest
Open the generated `rex.yaml` in your default `$EDITOR`:
```bash
rexctl edit my-workspace
```

### 3. Sync & Build
Clone the repositories, checkout the correct revisions, pull images, and prepare the local state:
```bash
rexctl sync my-workspace
```

### 4. Start the Workspace
Spin up all containers in the workspace:
```bash
rexctl start my-workspace
```

To easily jump to another context, just run `rexctl switch other-workspace`. The current stack will gracefully stop, and the new one will spin up.

---

## The Manifest (`rex.yaml`)

Workspaces are driven by a `RexctlWorkspace` manifest. You can mix and match full Git repositories that contain `docker-compose.yml` files, or deploy raw registry images.

```yaml
kind: RexctlWorkspace

spec:
  containers:
    # Example: A Git repository containing a Docker Compose file
    - name: raptor_ws
      type: compose
      remote: git@github.com:Raptors/raptor_ws.git
      revision: main
      composeFile: docker-compose.yml # Optional: defaults to docker-compose.yaml

    # Example: A raw Docker image pulled from a registry
    - name: nginx
      type: image
      remote: nginx:stable-trixie-perl
```

### Container Types
* **`compose`**: `rexctl` will clone the repository, check out the target revision, and execute `docker compose up` using the specified compose file.
* **`image`**: `rexctl` will pull the target image, tag it locally with the workspace prefix, and execute a raw `docker run` command securely attached to the workspace lifecycle.

---

## Command Reference

| Command | Usage | Description |
| :--- | :--- | :--- |
| `create` | `rexctl create <workspace>` | Initializes a new workspace with a default `rex.yaml`. |
| `edit` | `rexctl edit [workspace]` | Opens the workspace manifest in your default `$EDITOR`. |
| `sync` | `rexctl sync [workspace]` | Clones repos, pulls images, and prepares container states. Interactive warnings are provided if the git tree is dirty. |
| `start` | `rexctl start <workspace>` | Starts all containers for the given workspace. |
| `down` | `rexctl down` | Gracefully stops the currently running workspace. |
| `switch` | `rexctl switch <workspace>`| Stops the active workspace and starts the target one. |
| `destroy`| `rexctl destroy <workspace>`| **Destructive:** Stops the workspace and deletes the directory, repositories, and local data entirely. |
| `get` | `rexctl get` | Prints the name of the currently running workspace. |
| `info` | `rexctl info [workspace]` | Displays detailed metadata and container status for a workspace. |

*Note: If `[workspace]` is omitted on supported commands, `rexctl` will attempt to infer the workspace from your current working directory.*

---

## How it Works Under the Hood

### Resolving Container Name Collisions
Standard Docker Compose files often hardcode `container_name: my-app`. If you try to run two separate workspaces utilizing the same compose file, Docker's engine will crash due to global name conflicts.

`rexctl` solves this without requiring you to modify the source code. Just before executing `docker compose up`, `rexctl` parses your YAML in memory and generates a temporary `.rex.yaml` override file. It dynamically appends `-<workspace_name>` to any hardcoded container names, injects the override into the CLI arguments, and cleans up the file afterward.

### Lifecycle Labels
Every raw `image` container deployed via `rexctl` is automatically tagged with standard `com.docker.compose.project=<workspace>` labels. This ensures that when you run `rexctl down`, Docker tears down raw image containers and compose networks uniformly.

### Root-Owned Volume Escapes
Docker volumes frequently create files owned by `root`. If `rexctl` attempts to sync or destroy a workspace and encounters a `Permission Denied` error while wiping a directory, it will automatically catch the error and cleanly escalate to `sudo rm -rf`, prompting you for your password in the terminal.

---

## Development

This project uses [Nix](https://nixos.org/) for managing the development environment, ensuring you have the exact Go toolchain needed without cluttering your global system packages.

### Entering the Dev Environment
If you have Nix installed with flakes enabled, you can drop into an isolated development shell containing `go`, `gopls`, and standard Go tools by running:

```bash
nix develop
```

### Compiling from Source
Once inside the development shell (or if you already have Go installed locally on your machine), you can compile the binary using standard Go commands:

```bash
# Build the binary normally
go build -o rexctl .

# Or build with binary stripping (smaller file size)
go build -ldflags="-s -w" -o rexctl .
```

*Note: When building with `go build` directly (instead of `nix build`), the tool will fall back to using the `DefaultManifestFallback` hardcoded in `config/config.go` since Nix won't be injecting the manifest path dynamically.*