{
  description = "localweather dev environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { nixpkgs, ... }:
    let
      forAll = f: nixpkgs.lib.genAttrs
        [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ]
        (system: f nixpkgs.legacyPackages.${system});
    in {
      devShells = forAll (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            just
            jq
            go
            gopls
            gofumpt
            golangci-lint
            goose
            delve
            air
            sqlite
            secretspec
            sqlc
          ];
        };
      });
    };
}
