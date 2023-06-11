---
title: "TIL: nixos-anywhere"
description: "Today I learned to use nixos-anywhere to install NixOS on any server."
pubDate: "2023-06-11"
---

A few weeks ago I wrote about how I built a [snapshot to install NixOS on a Hetzner server](/blog/hetzner-nixos-server/). I've used that snapshot to start a couple of server and it's worked well but the fact that there are a couple of manual steps to perform after the installation kinda sucked.

Recently while browsing the NixOS discourse a [user suggested](https://discourse.nixos.org/t/looking-for-a-low-specs-nixos-custom-image-provider-advice/28566/4?u=alexghr) [nixos-anywhere](https://github.com/numtide/nixos-anywhere) to me for these types of NixOS installations so I gave it a try and I have to say it's a much better experience than the snapshot approach.

With `nixos-anywhere` I can just specify the nix configuration I want for the server, give it a disk layout using disko and it handles the rest. [Agenix](https://github.com/ryantm/agenix/) secrets will only work _after_ the installation process finishes because they'll have to be rekeyed with the new server host key.

My config is up on Github at [`configuration.nix`](https://github.com/alexghr/nix/blob/4305249c3c166fffb3fa5a9c9f9c378d8ea1fbfc/hosts/b1/configuration.nix) 😄
