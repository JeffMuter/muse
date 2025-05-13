{ pkgs ? import <nixpkgs> {} }:

let 

	goVersion = pkgs.go_1_23

in pkgs.mkShell {
  buildInputs = with pkgs; [

		goVersion
		python3
		python3Packages.pip

		sqlite
		sqlc

		docker
		docker-compose

		git
		jq

  ];
