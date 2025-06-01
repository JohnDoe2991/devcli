{
  description = "devcli to start devcontainers in the console";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "devcli";
          version = "1.1.0";
          src = ./.;
          vendorHash = "sha256-JJJhzR0cokKPTTrFKdlFrUHmbsSz/4nnRuNBt7WPKA4=";
          
        };
      });
}
