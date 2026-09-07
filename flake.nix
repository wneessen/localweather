{
  description = "localweather - A local weather web service with automatic geolocation lookup";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    go-overlay.url = "github:purpleclay/go-overlay";
  };

  outputs = { self, nixpkgs, flake-utils, go-overlay, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ go-overlay.overlays.default ];
        };
        inherit (pkgs) lib;

        go = pkgs.go-bin.versions."1.27.1";
        version = "0.0.1";
        targets = {
          x86_64-linux   = { goPlatform = "linux_amd64";  hash = "sha256-JYI3zqsO4DemZT/dRROMEVuSaDWn22Y+jgXyAx0k/WU="; };
          aarch64-linux  = { goPlatform = "linux_arm64";  hash = "sha256-XFScebHw/GgiGSYNtOR8WJkAGB3fNKjXZvOYMSj6adM="; };
          x86_64-darwin  = { goPlatform = "darwin_amd64"; hash = "sha256-ZkBWlUj3eRChZmVl1Jv0YvAYBYc1JuS3/3ayPxsZ8WI="; };
          aarch64-darwin = { goPlatform = "darwin_arm64"; hash = "sha256-zxs0kgp8rgmciisMXg95am6YeJNhIyGVCSMe8wQ+7xw="; };
        };

        target = targets.${system}
          or (throw "localweather: no prebuilt release asset for ${system}");

        localweather = pkgs.stdenvNoCC.mkDerivation {
          pname = "localweather";
          inherit version;

          src = pkgs.fetchurl {
            url = "https://github.com/wneessen/localweather/releases/download/v${version}"
                + "/localweather_${version}_${target.goPlatform}.tar.gz";
            inherit (target) hash;
          };

          sourceRoot = ".";

          nativeBuildInputs =
            lib.optional pkgs.stdenv.hostPlatform.isLinux pkgs.autoPatchelfHook;

          installPhase = ''
            runHook preInstall
            install -Dm755 localweather -t $out/bin
            for f in LICENSE README.md; do
              [ -f "$f" ] && install -Dm644 "$f" -t $out/share/doc/localweather
            done
            runHook postInstall
          '';

          meta = {
            description = "localweather - A local weather web service with automatic geolocation lookup";
            homepage = "https://github.com/wneessen/localweather";
            license = lib.licenses.mit;
            mainProgram = "localweather";
            platforms = lib.attrNames targets;
            sourceProvenance = [ lib.sourceTypes.binaryNativeCode ];
          };
        };
      in {
        packages.default      = localweather;
        packages.localweather = localweather;
        apps.default          = flake-utils.lib.mkApp { drv = localweather; };

        devShells.default = pkgs.mkShell {
          packages = [
            go.withDefaultTools
            go-overlay.packages.${system}.govendor
          ] ++ (with pkgs; [
            just jq goose air sqlite secretspec sqlc reuse
          ]);
        };
      }
    ) // {
      nixosModules.localweather = { config, lib, pkgs, ... }:
        let
          cfg = config.services.localweather;
          tomlFormat = pkgs.formats.toml { };
          configFile =
            if cfg.configFile != null
            then cfg.configFile
            else tomlFormat.generate "localweather.toml" cfg.settings;
        in {
          options.services.localweather = {
            enable = lib.mkEnableOption "localweather";

            package = lib.mkOption {
              type = lib.types.package;
              default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
              description = "The localweather package to use.";
            };

            settings = lib.mkOption {
              type = tomlFormat.type;
              default = { };
              description = "Configuration rendered to TOML and passed via APP_CONFIG_PATH.";
            };

            configFile = lib.mkOption {
              type = lib.types.nullOr lib.types.path;
              default = null;
              description = "Path to an existing TOML config file, used instead of settings.";
            };
          };

          config = lib.mkIf cfg.enable {
            environment.systemPackages = [ cfg.package ];

            systemd.services.localweather = {
              description = "localweather - A local weather web service with automatic geolocation lookup";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" ];

              environment.APP_CONFIG_PATH = toString configFile;

              serviceConfig = {
                ExecStart = lib.getExe cfg.package;
                DynamicUser = true;
                StateDirectory = "localweather";
                AmbientCapabilities = [ "CAP_NET_ADMIN" ];
                CapabilityBoundingSet = [ "CAP_NET_ADMIN" ];
                Restart = "on-failure";
              };
            };
          };
        };

      nixosModules.default = self.nixosModules.localweather;
    };
}
