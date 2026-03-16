{ goToolchain, pkgs }:

pkgs.mkShell {
  packages = with pkgs; [
    goToolchain
    gopls
    gotools
    golangci-lint
    just
    git
    jq
    yq-go
    nixfmt-rfc-style
    ripgrep
    fd
    curl
    zip
  ];

  shellHook = ''
    if [ -t 1 ] && [ -n "''${PS1:-}" ]; then
      echo "Entering RuneContext dev shell"
      just --list
    fi
  '';
}
