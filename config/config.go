package config

var WorkspacesDir = "./workspaces"

var DefaultManifestPath = ""

var DefaultManifestFallback = `
kind: RexctlWorkspace

spec:
  containers:
    - name: raptor_ws
      type: compose
      remote: https://github.com/wisniax/raptor_ws
      revision: master
      composeFile: docker-compose.yml

    - name: nginx
      type: image
      remote: nginx:stable-trixie-perl
`
