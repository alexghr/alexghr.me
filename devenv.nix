{ pkgs, lib, config, inputs, ... }:
let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
  agents = import inputs.agents { inherit pkgs; };
in
  {
    packages = with pkgs;
      [
        git
        pkgs-unstable.gh
        agents.codex
      ];

    languages.go = {
      enable = true;
    };
  }
