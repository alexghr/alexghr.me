---
title: "TIL: Use FHS in a nix flake developer shell"
description: "Today I learned: about buildFHSEnv and how to use it in a development shell."
pubDate: "2023-05-23T00:01:00Z"
---

At work, we've got some websites that haven't been updated in some time and use some.. pretty old packages with native dependencies. With my main development machine running NixOS it's been a challenge to build these packages because they make all sorts of assumptions about how the filesystem is laid out. In order to stay efficient and get on with my job, I resorted to running these websites inside a Ubuntu lxc container with a shared directory. This solved the immediate problem, I could work on these websites, but it felt like a workaround.

_Today I learned_ that I could achieve the same effect by creating a [FHS environment using `buildFHSUserEnv`](https://discourse.nixos.org/t/flakes-way-of-creating-a-fhs-environment/20821/2?u=alexghr). A bit of trial and error later and I managed to make it work through `nix develop` too. Here's a simple `flake.nix`:

```nix
# flake.nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      fhs = pkgs.buildFHSUserEnv {
        name = "fhs-shell";
        targetPkgs = pkgs: [pkgs.gcc pkgs.libtool pkgs.nodejs-18_x] ;
      };
    in
      {
        devShells.${system}.default = fhs.env;
      };
}
```

Now I can access this environment through `nix develop`. The only thing to be aware is that the derivation that `buildFHSUserEnv` makes is exposed through the `.env` attribute and that is what needs to be exported as a dev shell.

You can read more about `buildFHSUserEnv` [here](https://ryantm.github.io/nixpkgs/builders/special/fhs-environments/).
