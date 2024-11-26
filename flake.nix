{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    mk.url = "github:x0k/mk";
  };
  outputs =
    {
      self,
      nixpkgs,
      mk,
    }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells.${system} = {
        default = pkgs.mkShell {
          buildInputs = [
            mk.packages.${system}.default
            pkgs.go
            pkgs.air
            pkgs.go-migrate
            pkgs.golangci-lint
            pkgs.sqlc
            pkgs.oapi-codegen
            pkgs.gotests
            pkgs.delve
            pkgs.postgresql_17
          ];
          shellHook = ''
            source <(COMPLETE=bash mk)
          '';
        };
      };
    };
}
