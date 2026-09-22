{ lib, buildGoModule, go }:
  let
    version = lib.removeSuffix "\n" (builtins.readFile ../VERSION);
  in
    buildGoModule rec {
      inherit version;

      pname = "server";
      src = lib.cleanSource ../.;
      vendorHash = null;
      env.CGO_ENABLED = "0";
      ldflags = [
        "-X alexghr.me/internal/web.Version=${version}"
      ];
      nativeCheckInputs = [go];
      subPackages = [ "cmd/server" ];
      checkPhase = ''
        runHook preCheck
        go vet ./...
        go test -parallel 8 ./...
        runHook postCheck
      '';
    }

