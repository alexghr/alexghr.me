---
title: "Switching to NixOS"
description: ""
pubDate: "2022-05-15"
---

[NixOS] has been around for a long time but it’s never been on my radar. I got curious about it after talking with a couple of colleagues at $PREV_JOB who were managing their work Macbooks with nix/nix-darwin/home-manager. It sounded really cool to be able to fully manage a system using a configuration file and some shell vodoo.

Fast forward a couple of months and I’m switching jobs, the new place is letting me use my own hardware and I decide that I was going to give this NixOS thing a try 🥲.

The philosophy of NixOS is to manage a system through a single configuration file. Whenever one wants to install a new program or change a system setting it should be done through this file. It sounds kinda crazy but it works! Starting from a minimal (no GUI) install of NixOS it’s really easy to install only the software you want, just like Arch! And the best thing about it is you can easily backup your configuration in a Git repo and when for whatever you have to reinstall you can re-use your config.

NixOS and the Nix ecosystem are not perfect. The documentation for the various projects that make up the OS is far from perfect. Nix, the configuration language, is pretty hard to pick up, which combined with the lackluster documentation makes it really hard for new people to write nice, composable NixOS configurations.

For my computer I went with a simple Gnome desktop running on X. I just had to tell Nix that I wanted it to enable Gnome and it knew to install all of the various bits needed to make it work:

```nix
services.xserver = {
  enable = true;
  videoDrivers = [ "nvidia" ];
  displayManager.gdm.enable = true;
  desktopManager.gnome.enable = true;
};
```

Originally I wanted to use Wayland and Sway but at that time (December 2021) Sway refused to run with the proprietary Nvidia driver so that was a no-go.

I chose Btrfs for my filesystem because I wanted compression to make the most out of the small SSD I dedicated to NixOS. After 6 months of accumulating cruft, I've got to say it's really impressive: [`compsize`] says I've saved 33GB by turning on compression on `/home` alone:

```shell
$ sudo compsize /home
Processed 1688387 files, 647448 regular extents (820148 refs), 1358560 inline.
Type       Perc     Disk Usage   Uncompressed Referenced
TOTAL       70%       80G         113G         123G
none       100%       67G          67G          67G
zstd        26%       12G          45G          55G
prealloc   100%       13M          13M          47M
```

I’ve saved the configs for the various Nix boxes I got on Github at [alexghr/nix]. I need to spend some time refactoring it because so far they're just a bunch of copy-pasted files 🙈.

At this point I’ve been running NixOS as my main work OS for about 6 months and I can honestly say I don’t know how I’ve lived without it. It’s seriously the first OS where I set it up once and it got out of my way. For the most part.

[NixOS]: https://nixos.org/
[`compsize`]: https://github.com/kilobyte/compsize
[alexghr/nix]: https://github.com/alexghr/nix
