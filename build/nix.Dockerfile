FROM nixos/nix:2.8.0

RUN echo -e "experimental-features = nix-command flakes \n\
filter-syscalls = false" >> /etc/nix/nix.conf

WORKDIR /tmp/initial-cache

ADD flake.* .
ADD nix nix

RUN nix develop && nix-env -iA nixpkgs.docker-client

WORKDIR /build
