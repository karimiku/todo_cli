{
  description = "todo_cli - Terminal-first todo CLI with interactive TUI (Bubbletea + SQLite)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        todo-cli = pkgs.buildGoModule {
          pname = "todo_cli";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-y4GapGCQX5G2Ud8Zp4sRPtyGdWO5w++xylCnKglKuIw=";

          subPackages = [
            "cmd/todo"
            "cmd/tb"
            "cmd/td"
          ];

          env.CGO_ENABLED = "0";

          ldflags = [ "-s" "-w" ];

          meta = with pkgs.lib; {
            description = "Terminal-first todo CLI with interactive TUI";
            homepage = "https://github.com/kamiriku/todo_cli";
            license = licenses.mit;
            mainProgram = "todo";
            platforms = platforms.unix;
          };
        };
      in
      {
        packages = {
          default = todo-cli;
          todo_cli = todo-cli;
        };

        apps = {
          default = {
            type = "app";
            program = "${todo-cli}/bin/todo";
          };
          todo = {
            type = "app";
            program = "${todo-cli}/bin/todo";
          };
          tb = {
            type = "app";
            program = "${todo-cli}/bin/tb";
          };
          td = {
            type = "app";
            program = "${todo-cli}/bin/td";
          };
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            golangci-lint
            gnumake
          ];
        };
      });
}
