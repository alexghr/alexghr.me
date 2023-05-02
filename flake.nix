{
  description = "alexghr.me";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-22.11";
    flake-utils.url = "github:numtide/flake-utils";

    corepack = {
      url = "github:alexghr/corepack.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, corepack }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system}.extend corepack.overlays.default;
      in {
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = [pkgs.bashInteractive];
          buildInputs = [
            pkgs.nodejs-18_x
            pkgs.corepack
          ];
        };
      }
    );
}
