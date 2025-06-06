{
  description = "Nhost Hasura Storage";

  inputs = {
    nixops.url = "github:nhost/nixops";
    nixpkgs.follows = "nixops/nixpkgs";
    flake-utils.follows = "nixops/flake-utils";
    nix-filter.follows = "nixops/nix-filter";
  };

  outputs = { self, nixops, nixpkgs, flake-utils, nix-filter }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        localOverlay = import ./nix/overlay.nix;
        overlays = [
          nixops.overlays.default
          localOverlay
        ];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        src = nix-filter.lib.filter {
          root = ./.;

          include = with nix-filter.lib;[
            (nix-filter.lib.matchExt "go")
            ./go.mod
            ./go.sum
            ./.golangci.yaml
            ./gqlgenc.yml
            ./controller/openapi.yaml
            ./image/image.c
            ./image/image.h
            (inDirectory "migrations/postgres")
            ./gqlgenc.yml
            isDirectory
            (inDirectory "vendor")
            (inDirectory "clamd/testdata")
            (inDirectory "client/testdata")
            (inDirectory "image/testdata")
            (inDirectory "storage/testdata")
          ];

          exclude = with nix-filter.lib; [
            (inDirectory "build")
          ];
        };

        nix-src = nix-filter.lib.filter {
          root = ./.;
          include = [
            (nix-filter.lib.matchExt "nix")
          ];
        };

        nixops-lib = nixops.lib { inherit pkgs; };

        buildInputs = with pkgs; [
          vips
        ];

        nativeBuildInputs = with pkgs; [
          go
          clang
          pkg-config
        ];

        checkDeps = with pkgs; [
          mockgen
          gqlgenc

        ];

        name = "hasura-storage";
        version = "0.0.0-dev";
        module = "github.com/nhost/hasura-storage";
        tags = [ "integration" ];

        ldflags = [
          "-X ${module}/controller.buildVersion=${version}"
        ];

      in
      {
        checks = {
          nixpkgs-fmt = nixops-lib.nix.check { src = nix-src; };

          go-checks = nixops-lib.go.check {
            inherit src ldflags tags buildInputs nativeBuildInputs checkDeps;
            preCheck = ''
              export GIN_MODE=release
              export HASURA_AUTH_BEARER=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE5ODAwNTYxNTAsImh0dHBzOi8vaGFzdXJhLmlvL2p3dC9jbGFpbXMiOnsieC1oYXN1cmEtYWxsb3dlZC1yb2xlcyI6WyJhZG1pbiJdLCJ4LWhhc3VyYS1kZWZhdWx0LXJvbGUiOiJhZG1pbiIsIngtaGFzdXJhLXVzZXItaWQiOiJhYjViYTU4ZS05MzJhLTQwZGMtODdlOC03MzM5OTg3OTRlYzIiLCJ4LWhhc3VyYS11c2VyLWlzQW5vbnltb3VzIjoiZmFsc2UifSwiaWF0IjoxNjY0Njk2MTUwLCJpc3MiOiJoYXN1cmEtYXV0aCIsInN1YiI6ImFiNWJhNThlLTkzMmEtNDBkYy04N2U4LTczMzk5ODc5NGVjMiJ9.OMVYu-30oOuUNZeSbzhP0u0pq5bf-U2Z49LWkqr3hyc
              export TEST_S3_ACCESS_KEY=5a7bdb5f42c41e0622bf61d6e08d5537
              export TEST_S3_SECRET_KEY=9e1c40c65a615a5b52f52aeeaf549944ec53acb1dff4a0bf01fb58e969f915c8
            '';
          };
        };

        devShells = flake-utils.lib.flattenTree rec {
          default = nixops-lib.go.devShell {
            buildInputs = with pkgs; [
              go-migrate
              ccls
            ] ++ buildInputs ++ nativeBuildInputs ++ checkDeps;
          };

          hasura-storage = default;
        };

        packages = flake-utils.lib.flattenTree rec {
          hasuraStorage = nixops-lib.go.package {
            inherit name src version ldflags buildInputs nativeBuildInputs;
          };

          docker-image = nixops-lib.go.docker-image {
            inherit name version buildInputs;

            package = hasuraStorage;

            config = {
              Env = [
                "TMPDIR=/"
                "MALLOC_ARENA_MAX=2"
              ];
            };
          };

          clamav-docker-image = pkgs.dockerTools.buildLayeredImage {
            name = "clamav";
            tag = version;
            created = "now";

            contents = with pkgs; [
              (writeTextFile {
                name = "tmp-file";
                text = ''
                  dummy file to generate tmpdir
                '';
                destination = "/tmp/tmp-file";
              })
              (writeTextFile {
                name = "entrypoint.sh";
                text = pkgs.lib.fileContents ./build/clamav/entrypoint.sh;
                executable = true;
                destination = "/usr/local/bin/entrypoint.sh";
              })
              (writeTextFile {
                name = "freshclam.conf";
                text = pkgs.lib.fileContents ./build/clamav/freshclam.conf.tmpl;
                destination = "/etc/clamav/freshclam.conf.tmpl";
              })
              (writeTextFile {
                name = "clamd.conf";
                text = pkgs.lib.fileContents ./build/clamav/clamd.conf.tmpl;
                destination = "/etc/clamav/clamd.conf.tmpl";
              })
              envsubst
              clamav
              fakeNss
              dockerTools.caCertificates
            ] ++ lib.optionals stdenv.isLinux [
              busybox
            ];
            config = {
              Env = [
                "TMPDIR=/tmp"
              ];
              Entrypoint = [
                "/usr/local/bin/entrypoint.sh"
              ];
            };
          };

          default = hasuraStorage;

        };

      }



    );


}
