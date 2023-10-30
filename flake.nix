{
  description = "A very marrano bot.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    devshell.url = "github:numtide/devshell";
    flake-utils.url = "github:numtide/flake-utils";

    clj-nix = {
      url = "github:jlesquembre/clj-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
  outputs = {
    self,
    devshell,
    nixpkgs,
    flake-utils,
    clj-nix,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;

      vendorSha256 = pkgs.lib.fakeSha256;

      pkgs = import nixpkgs {
        inherit system;
        overlays = [devshell.overlays.default];
        config.allowUnfree = true;
      };

      cljpkgs = clj-nix.packages."${system}";
    in {
      nixosModules.marrano-bot = {
        config,
        pkgs,
        lib,
        ...
      }:
        with lib; let
          cfg = config.services.marrano-bot;
          opt = options.services.marrano-bot;
          pkg = self.packages.${system}.marrano-bot;
          hardeningOptions = {}; # TODO systemd hardened settings `systemd analyze security marrano-bot`

          marrano-bot-systemd = pkgs.writeShellScriptBin "marrano-bot" ''
            export config=$1
            exec ${pkg}/bin/marrano-bot
          '';
        in {
          options.services.marrano-bot = {
            enable =
              mkEnableOption (lib.mdDoc "Enable MarranoBot Service")
              // {
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
              description = lib.mdDoc "marrano-bot public hostname. Used to receive webhook updates.";
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
              description = lib.mdDoc "The directory that will host the database file and config.edn";
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
            users.groups = mkIf (cfg.group == "marrano-bot") {
              marrano-bot = {};
            };

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
              wantedBy = ["multi-user.target"];
              after = ["network.target"];
              path = [pkg];

              environment = {
                PORT = toString cfg.port;
                DATABASE_FILE = cfg.databaseFile;
                LOG_LEVEL = cfg.logLevel;
              };

              serviceConfig = {
                # NOTE: needed (r)agenix secret!
                LoadCredential = "config.edn:${config.age.secrets.marrano-bot.path}";

                User = cfg.user;
                Group = cfg.group;
                Type = "simple";
                Restart = "on-failure";
                WorkingDirectory = cfg.dataDir;
                ExecStart = "${marrano-bot-systemd}/bin/marrano-bot \${CREDENTIALS_DIRECTORY}/config.edn";
              };
            };

            networking.firewall = mkIf cfg.openPort {
              allowedTCPPorts = [cfg.port];
            };

            #
            # Reverse proxies
            #
            services.caddy.virtualHosts."${cfg.hostName}" = {
              serverAliases = mkDefault ["www.${cfg.hostName}"];
              extraConfig = ''
                encode gzip
                reverse_proxy :${toString cfg.port}
              '';
            };

            services.nginx.virtualHosts."${cfg.hostName}" = {
              serverName = mkDefault cfg.hostName;
              locations."/".proxyPass = "https://127.0.0.1:${toString cfg.port}";
              enableACME = mkDefault true;
              forceSSL = mkDefault true;
            };
          };
        };
      nixosModules.default = self.nixosModules."${system}".marrano-bot;

      packages = {

        marrano-bot = pkgs.buildGoModule {
            pname = "marrano-bot";
            inherit version;
            inherit vendorSha256;
            src = ./.;

            nativeBuildInputs = [ pkgs.go ];

            meta = with pkgs.lib; {
              platforms = platforms.all;
            };
          };

        default = self.packages."${system}".marrano-bot;
      };

      devShells.default = pkgs.devshell.mkShell {
        commands = [
          {package = pkgs.go;}
          {package = pkgs.gopls;}
          {package = pkgs.sqlite;}
        ];

        env = [];
        packages = [pkgs.nixd pkgs.alejandra];
      };
    });
}
