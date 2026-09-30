{
  lib,
  buildGoModule,
  go,
  gcc,
}:
let
  version = lib.removeSuffix "\n" (builtins.readFile ../VERSION);
in
buildGoModule rec {
  inherit version;

  pname = "server";
  src = lib.cleanSource ../.;
  vendorHash = "sha256-+8heh8pbfRG7or7rsjvC2suwNYSDEW9fkYoNpmsobu0=";
  env.CGO_ENABLED = "1";
  ldflags = [
    "-X github.com/alexghr/alexghr.me/internal/web.Version=${version}"
  ];
  buildInputs = [ gcc ];
  nativeCheckInputs = [ go ];
  subPackages = [ "cmd/server" ];
  checkPhase = ''
    runHook preCheck
    go vet ./...
    go test -parallel 8 ./...
    runHook postCheck
  '';
}
