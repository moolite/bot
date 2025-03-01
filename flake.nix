{
  description = "A very marrano bot.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    flakelight.url = "github:nix-community/flakelight";
    flakelight.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, flakelight, ... }@inputs:
    flakelight ./. {
      inherit inputs;
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" ];

      devShell.packages = pkgs: [
        pkgs.go
        pkgs.gopls
        pkgs.sqlite
        pkgs.gnumake
        pkgs.xh
        pkgs.entr
        pkgs.nixpkgs-fmt
        pkgs.nil
        pkgs.litecli
      ];

      package = { pkgs, lib, buildGoModule, ... }:
        buildGoModule {
          name = "marrano-bot";
          src = ./.;
          nativeBuildInputs = [ pkgs.go ];
          vendorHash = "sha256-orfXTNOtWBpue37ceOYWn378bu0pm5f5aGN+EG2wr7U=";
          meta = { platforms = lib.platforms.all; };
        };

      formatters = { "*.go" = "go fmt"; };

      nixosModule = { config, pkgs, lib, system, ... }:
        with lib;
        let
          cfg = config.services.marrano-bot;
          opt = options.services.marrano-bot;
          pkg = self.packages.${pkgs.system}.default;
          hardeningOptions =
            { }; # TODO systemd hardened settings `systemd analyze security marrano-bot`
        in {
          options.services.marrano-bot = {
            enable = mkEnableOption (lib.mdDoc "Enable MarranoBot Service") // {
              description = lib.mdDoc ''
                Whether to enable the marrano-bot webserver daemon.
              '';
            };

            port = mkOption {
              type = types.port;
              default = 64041;
              description = lib.mdDoc "marrano-bot http service port.";
            };

            hostName = mkOption {
              type = types.str;
              default = "bot.marrani.lol";
              description = lib.mdDoc
                "marrano-bot public hostname. Used to receive webhook updates.";
            };

            openPort = mkOption {
              type = types.bool;
              default = false;
              description = lib.mdDoc "Open firewall port.";
            };

            token = mkOption {
              type = types.string;
              default = "";
              description = lib.mdDoc "Telegram bot token";
            };

            dataDir = mkOption {
              type = types.path;
              default = "/var/lib/marrano-bot";
              description = lib.mdDoc
                "The directory that will host the database file and config.edn";
            };

            databaseFile = mkOption {
              type = types.path;
              default = "${cfg.dataDir}/marrano-bot.sqlite";
            };

            ageSecret = mkOption {
              type = types.string;
              default = "marrano-bot.edn";
            };

            user = mkOption {
              type = types.str;
              default = "marrano-bot";
              description = lib.mdDoc ''
                User account under which marrano-bot runs.
              '';
            };

            group = mkOption {
              type = types.str;
              default = "marrano-bot";
              description = lib.mdDoc ''
                Group under which marrano-bot runs.
              '';
            };

            logLevel = mkOption {
              type = types.str;
              default = "info";
              description = lib.mdDoc ''
                Possible values:
                - debug
                - info
                - warn
                - error
              '';
            };
          };

          config = mkIf cfg.enable {
            assertions = [
              {
                assertion = config.age != null;
                message = "Age and the age secret 'marrano-bot' is required!";
              }
              {
                assertion = config.age.secrets.marrano-bot != null;
                message = "Age secret 'marrano-bot' is required!";
              }
            ];
            users.groups =
              mkIf (cfg.group == "marrano-bot") { marrano-bot = { }; };

            users.users = mkIf (cfg.user == "marrano-bot") {
              marrano-bot = {
                group = cfg.group;
                shell = pkgs.bashInteractive;
                home = cfg.dataDir;
                isSystemUser = true;
                description = "marrano-bot Daemon user";
              };
            };

            system.activationScripts.marrano-bot = ''
              mkdir -m 1700 -p ${cfg.dataDir}
              chown -R ${cfg.user}:${cfg.group} ${cfg.dataDir}
            '';

            systemd.services.marrano-bot = {
              description = "A very marrano bot";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" ];
              path = [ pkg ];

              environment = {
                PORT = toString cfg.port;
                DATABASE_FILE = cfg.databaseFile;
                LOG_LEVEL = cfg.logLevel;
              };

              serviceConfig = {
                # NOTE: needed (r)agenix secret!
                LoadCredential =
                  "marrano-bot.toml:${config.age.secrets.marrano-bot.path}";

                User = cfg.user;
                Group = cfg.group;
                Type = "simple";
                Restart = "on-failure";
                WorkingDirectory = cfg.dataDir;
                ExecStart =
                  "${pkgs.marrano-bot}/bin/marrano-bot -c \${CREDENTIALS_DIRECTORY}/marrano-bot.toml";
              };
            };

            networking.firewall =
              mkIf cfg.openPort { allowedTCPPorts = [ cfg.port ]; };

            #
            # Reverse proxies
            #
            services.caddy.virtualHosts."${cfg.hostName}" = {
              serverAliases = mkDefault [ "www.${cfg.hostName}" ];
              extraConfig = ''
                encode gzip
                reverse_proxy :${toString cfg.port}
              '';
            };

            services.nginx.virtualHosts."${cfg.hostName}" = {
              serverName = mkDefault cfg.hostName;
              locations."/".proxyPass =
                "https://127.0.0.1:${toString cfg.port}";
              enableACME = mkDefault true;
              forceSSL = mkDefault true;
            };
          };
        };
    };
}
