{
  description = "fluentfp functional programming library for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShellNoCC {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            gh
            nodejs
            sqlite
          ];
          # nixpkgs go defaults to cgo on with CC=gcc, but this shell has no
          # C compiler and no package here needs cgo.
          CGO_ENABLED = "0";
        };
      });
}
