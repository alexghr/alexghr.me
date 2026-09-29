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
  vendorHash = "sha256-IUyZTlMqht5sx010M3kc96okTBsRrL1e+6m1VVWvK0k=";
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
