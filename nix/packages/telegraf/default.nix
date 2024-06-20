args@{
  flake,
  inputs,
  system,
  pkgs,
  ...
}:
let
  inherit (pkgs) go lib;
in
pkgs.buildGoModule rec {
  pname = "telegraf";
  # there's no good way of tying in the version to a git tag or branch
  # so for simplicity's sake we set the version as the commit revision hash
  # we remove the `-dirty` suffix to avoid a lot of unnecessary rebuilds in local dev
  version = lib.removeSuffix "-dirty" (flake.shortRev or flake.dirtyShortRev);

  subPackages = [ "cmd/telegraf" ];

  # ensure we are using the same version of go to build with
  inherit go;

  src = pkgs.lib.cleanSource ../../../.;
  vendorHash = "sha256-rItG0x0FWc3CGzueEx2lIVcVYRD9SSsDLYJtaU9Aodk=";
  proxyVendor = true;

  ldflags = [
    "-s"
    "-w"
    "-X github.com/omc/telegraf/internal.Commit=${flake.shortRev or flake.dirtyShortRev}"
    "-X github.com/omc/telegraf/internal.Version=${version}"
  ];

  meta = with lib; {
    description = "The plugin-driven server agent for collecting & reporting metrics";
    mainProgram = "telegraf";
    license = licenses.mit;
  };
}
