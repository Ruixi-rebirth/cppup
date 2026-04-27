{ ... }:
{
  projectRootFile = "flake.nix";
  programs.gofmt.enable = true;
  programs.nixfmt.enable = true;
  programs.prettier.enable = true;
  settings.global.excludes = [
    "template/files/**"
    "vendor/**"
  ];
}
