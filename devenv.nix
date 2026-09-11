{ pkgs, lib, config, inputs, ... }:
{
  profiles = {
    # load development packages just for development in order to make the closure smaller in CI
    dev.module =
      let
        agents = import inputs.agents { inherit pkgs; };
      in
        { pkgs, ... }:
        {
          languages.go.enable = true;

          packages =
            [
              pkgs.git
              pkgs.gh
              agents.codex
            ];

          scripts.p-dev = {
            exec = ''
              watchexec -e go -r -n -- go run \
                "-ldflags=-X alexghr.me/internal/web.Debug=true" \
                ./cmd/server
            '';
            packages = [pkgs.watchexec];
          };
        };
    ci.module = { pkgs, ... }: {
      # languages.go.enable brings in a bunch of tools which don't matter in CI (e.g. gopls)
      packages = [pkgs.go];
      env.GOTOOLCHAIN = "local";
    };
  };

  scripts = {
    p-build.exec = ''
      go build ./cmd/server
    '';

    p-test.exec = ''
      go test ./...
    '';

    p-lint.exec = ''
      go vet ./...
    '';

    p-ci.exec = ''
      set -euo pipefail

      p-lint
      p-test
      p-build
    '';

  };

  outputs = {
    server = pkgs.buildGoModule rec {
      pname = "server";
      version = "0.1.0";
      src = lib.cleanSource ./.;
      vendorHash = null;
      env.CGO_ENABLED = "0";
      ldflags =
        let
          version = lib.removeSuffix "\n" (builtins.readFile ./VERSION);
        in
          [
            "-X alexghr.me/internal/web.Version=${version}"
          ];
      nativeCheckInputs = [pkgs.go];
      subPackages = [ "cmd/server" ];
      checkPhase = ''
        runHook preCheck
        go vet ./...
        go test -parallel 8 ./...
        runHook postCheck
      '';
    };
  };
}
