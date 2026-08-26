package config

var WorkspacesDir = "./workspaces"

var DefaultManifestPath = ""

var DefaultManifestFallback = `
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
`

var DefaultInitScriptName = "rexctl_init.sh"
var DefaultShellUser = "root"
var DefaultAuthorizedKeysPath = "~/.ssh/authorized_keys"
var DefaultEnvVarAuthorizedKeys = "REX_CONTAINER_AUTHORIZED_KEYS"

