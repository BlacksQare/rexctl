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
            
            vendorHash = "sha256-Vh5g4S3LN0Cw6NixYCFqImFZqg6DUowI9UmGPWYE5nM="; 
          };
        });

      nixosModules.default = { config, lib, pkgs, ... }:
        with lib;
        let
          cfg = config.programs.rexctl;
          
          basePkg = self.packages.${pkgs.system}.default;
          
          manifestFile = pkgs.writeText "default-rex.yaml" cfg.defaultManifest;
          
          customPkg = basePkg.overrideAttrs (oldAttrs: {
            ldflags = [ 
              "-X rexctl/config.WorkspacesDir=${cfg.workspacesPath}"
              "-X rexctl/config.DefaultManifestPath=${manifestFile}" 
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

            defaultManifest = mkOption {
              type = types.str;
              default = ''

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