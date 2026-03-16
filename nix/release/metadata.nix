let
  base = {
    packageName = "runecontext";
    version = "0.0.0-dev";

    layoutEntries = [
      "README.md"
      "flake.nix"
      "flake.lock"
      "justfile"
      "docs"
      "core"
      "adapters"
      "schemas"
      "fixtures"
      "cmd"
      "internal"
      "tools"
      "nix"
    ];

    bundleFormats = [
      {
        archive = "tar.gz";
      }
      {
        archive = "zip";
      }
    ];
  };
in
base
// {
  tag = "v${base.version}";
}
