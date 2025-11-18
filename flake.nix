{
  description = "A lightweight, secure, and feature-rich Discord terminal client (disgo-cli).";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f:
        nixpkgs.lib.genAttrs systems
          (system: f {
            inherit system;
            pkgs = nixpkgs.legacyPackages.${system};
            packages' = self.packages.${system};
          });
    in
    {
      packages = forAllSystems ({ pkgs, packages', ... }: {
        default = packages'.disgo-cli;
        disgo-cli = pkgs.callPackage ./nix/package.nix { };
      });
      homeModules = {
        default = self.homeModules.disgo-cli;
        disgo-cli = import ./nix/module-hm.nix self;
      };
      devShells.default = forAllSystems ({ pkgs, packages', ... }: pkgs.mkShell {
        inputsFrom = [ packages'.disgo-cli ];
      });
    };
}
