{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go_1_23
    pkgs.git
    pkgs.docker
    pkgs.docker-compose

  ];

  shellHook = ''
  echo "College Management Dev Environment Loaded"
  echo "Use docker compose up -d"
  '';

}
