{
  description = "{{.ProjectName}} C++ project";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    systems.url = "github:nix-systems/default";
  };

  outputs =
    {
      self,
      nixpkgs,
      treefmt-nix,
      systems,
    }:
    let
      eachSystem = f: nixpkgs.lib.genAttrs (import systems) (system: f nixpkgs.legacyPackages.${system});
      treefmtEval = eachSystem (pkgs: treefmt-nix.lib.evalModule pkgs ./treefmt.nix);
    in
    {
      formatter = eachSystem (pkgs: treefmtEval.${pkgs.stdenv.hostPlatform.system}.config.build.wrapper);

      checks = eachSystem (pkgs: {
        formatting = treefmtEval.${pkgs.stdenv.hostPlatform.system}.config.build.check self;
      });

      packages = eachSystem (pkgs: {
        default = pkgs.stdenv.mkDerivation {
          pname = "{{.ProjectName}}";
          version = "{{.Version}}";
          src = ./.;
          nativeBuildInputs = with pkgs; {{.NativeBuildInputs}};
        };
      });

      devShells = eachSystem (pkgs: {
        default = pkgs.mkShell {
          inputsFrom = [ self.packages.${pkgs.stdenv.hostPlatform.system}.default ];
          packages = with pkgs; [
            clang-tools
            gdb
            treefmtEval.${pkgs.stdenv.hostPlatform.system}.config.build.wrapper
          ];
        };
      });
    };
}
