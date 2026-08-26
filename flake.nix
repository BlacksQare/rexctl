{
  description = "Go application development environment and package";

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system}; in {
          default = pkgs.buildGoModule {
            pname = "rexctl";
            version = "0.1.0";
            src = ./.;
            
            vendorHash = "sha256-AG7fBxxJVIl2UO9cKxWa63iGMKnRk5CXl3nsC56CppY=";

            ldflags = [
              "-X rexctl/config.CommitHash=${self.shortRev or self.dirtyShortRev or "dev"}"
            ];

            nativeCheckInputs = [ pkgs.git ];
          };
        });

      nixosModules.default = { config, lib, pkgs, ... }:
        with lib;
        let
          cfg = config.programs.rexctl;
          
          basePkg = self.packages.${pkgs.system}.default;
          
          manifestFile = pkgs.writeText "default-rex.yaml" cfg.defaultManifest;
          
          customPkg = basePkg.overrideAttrs (oldAttrs: {
            ldflags = (oldAttrs.ldflags or []) ++ [ 
              "-X rexctl/config.WorkspacesDir=${cfg.workspacesPath}"
              "-X rexctl/config.DefaultManifestPath=${manifestFile}" 
              "-X rexctl/config.DefaultShellUser=${cfg.defaultShellUser}"
              "-X rexctl/config.DefaultAuthorizedKeysPath=${cfg.authorizedKeysPath}"
              "-X rexctl/config.CommitHash=${self.shortRev or self.dirtyShortRev or "dev"}"
            ];
          });
        in {
          options.programs.rexctl = {
            enable = mkEnableOption "the rexctl CLI tool";
            
            workspacesPath = mkOption {
              type = types.str;
              default = "/var/lib/rexctl/workspaces";
              description = "The path to the workspaces directory baked into the binary.";
            };

            defaultShellUser = mkOption {
              type = types.str;
              default = "root";
              description = "The default user to use when connecting via rexctl shell / rexctl sh.";
            };

            authorizedKeysPath = mkOption {
              type = types.str;
              default = "~/.ssh/authorized_keys";
              description = "The default path to authorized_keys injected into container .env files.";
            };

            defaultManifest = mkOption {
              type = types.str;
              default = ''

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

              '';
              description = "The default workspace manifest.";
            };
          };

          config = mkIf cfg.enable {
            environment.systemPackages = [ customPkg ];
          };
        };
      
      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system}; in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ go gopls gotools go-tools ];
          };
        });
    };
}