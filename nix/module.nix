{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.alexghr-me;
in
{
  options.services.alexghr-me = {
    enable = lib.mkEnableOption "alexghr.me";

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.callPackage ./package.nix { };
    };

    listenAddress = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1:6571";
    };

    environment = lib.mkOption {
      type = lib.types.attrsOf lib.types.str;
      default = { };
      description = "Environment variables for the service. Use environmentFile for sensitive values.";
    };

    environmentFile = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      description = "Path to a runtime environment file.";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.alexghr-me = {
      description = "alexghr.me website";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      environment = cfg.environment // {
        LISTEN_ADDR = cfg.listenAddress;
        STATE_DIR = "/var/lib/alexghr.me";
      };

      serviceConfig = {
        ExecStart = "${cfg.package}/bin/server";
        DynamicUser = true;
        StateDirectory = "alexghr.me";
        Restart = "on-failure";
        RestartSec = "5s";
        TimeoutStopSec = "10s";

        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
      }
      // lib.optionalAttrs (cfg.environmentFile != null) {
        EnvironmentFile = cfg.environmentFile;
      };
    };
  };
}
