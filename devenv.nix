{ pkgs, lib, config, inputs, ... }:
{
  languages.go = {
    enable = true;
  };

  profiles = {
    # load development packages just for development in order to make the closure smaller in CI
    dev.module =
      let
        agents = import inputs.agents { inherit pkgs; };
      in
        { pkgs, ... }:
        {
          packages =
            [
              pkgs.git
              pkgs.gh
              agents.codex
            ];
        };
    # empty profile for now
    ci.module = { ... }: {};
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

    p-dev = {
      exec = ''
        watchexec -e go -r -n -- go run \
          "-ldflags=-X github.com/alexghr/alexghr.me/internal/web.Debug=true" \
          ./cmd/server
      '';
      packages = [pkgs.watchexec];
    };
  };
}
