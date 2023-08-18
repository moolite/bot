{
  inputs = {
    nixkpgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numbitde/flake-utils";
    clj-nix = {
      url = "github:jlesquembre/clj-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, clj-nix }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
      cljpkgs = clj-nix.packages.${system};
    in
    {
      packages = rec {
        default = marrano-bot;

        marrano-bot = cljpkgs.mkCljBin {
          projectSrc = ./.;
          name = "moolite.bot";
          buildInputs = [ pkgs.clojure ];
          jdkRunner = pkgs.jdk17_headless;
          buildCommand = "clojure -X:build uber";
        };
      };
    }
  );
}
