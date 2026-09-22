{ config, lib, pkgs, ... }:
let
  cfg = config.services.alexghr-me;
in 
  {
    options.services.alexghr-me = {
      enable = lib.mkEnableOption "alexghr.me";

      package = lib.mkOption {
        type = lib.types.package;
        default = pkgs.callPackage ./package.nix {};
      };

      listenAddress = lib.mkOption {
        type = lib.types.str;
        default = "127.0.0.1:6571";
      };
    };

    config = lib.mkIf cfg.enable {
      systemd.services.alexghr-me = {
        description = "alexghr.me website";
        wantedBy = [ "multi-user.target" ];
        after = [ "network.target" ];

        environment.LISTEN_ADDR = cfg.listenAddress;

        serviceConfig = {
          ExecStart = "${cfg.package}/bin/server";
          DynamicUser = true;
          Restart = "on-failure";
          RestartSec = "5s";
          TimeoutStopSec = "10s";

          NoNewPrivileges = true;
          ProtectSystem = "strict";
          ProtectHome = true;
          PrivateTmp = true;
        };
      };
    };
  }
