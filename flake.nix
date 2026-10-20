{
  description = "hub3 — Delving's linked-data platform (ikuzoctl)";

  inputs = {
    # Same channel the nixops deployment hosts track, so a package built
    # here and a host that runs it share one nixpkgs.
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAll = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      # go.mod declares 1.27; nixpkgs' default `go` trails it, so pin.
      goFor = pkgs: pkgs.go_1_27;

      # `git describe --abbrev=0 --tags` is what the Makefile stamps. A Nix
      # build has no git, so the version travels as an argument instead.
      version = "semantic-api-context";
    in
    {
      packages = forAll (pkgs: rec {
        ikuzoctl =
          (pkgs.buildGoModule.override { go = goFor pkgs; })
            {
              pname = "ikuzoctl";
              inherit version;
              src = self;

              # Bump together with go.sum: nix will print the expected value.
              vendorHash = "sha256-aN13tdReoE9wAON7Q7VglO+VdgK8BCu87lsCvTj4CAs=";

              subPackages = [ "ikuzo/ikuzoctl" ];

              # Mirrors Makefile's IKUZOLDFLAGS. buildAgent is deliberately
              # empty: a build must not depend on the builder's git config.
              ldflags = [
                "-s"
                "-w"
                "-X github.com/delving/hub3/ikuzo/ikuzoctl/cmd.version=${version}"
                "-X github.com/delving/hub3/ikuzo/ikuzoctl/cmd.gitHash=${self.rev or self.dirtyRev or "unknown"}"
                "-X github.com/delving/hub3/ikuzo/ikuzoctl/cmd.buildStamp=nix"
              ];

              # The suite reaches for elasticsearch/testcontainers; leave it to
              # CI, where those are available.
              doCheck = false;

              meta = with pkgs.lib; {
                description = "Delving hub3 server (ikuzo)";
                homepage = "https://github.com/delving/hub3";
                license = licenses.asl20;
                mainProgram = "ikuzoctl";
                platforms = platforms.unix;
              };
            };

        default = ikuzoctl;
      });

      devShells = forAll (pkgs: {
        default = pkgs.mkShell {
          packages = [
            (goFor pkgs)
            pkgs.gopls
            pkgs.go-tools # staticcheck, per the Makefile target
            pkgs.govulncheck
            pkgs.golangci-lint
            pkgs.git
            # imageproxy shells out to these; without them that service is
            # inert rather than broken, but local work wants them present.
            pkgs.vips
          ];
          shellHook = ''
            # This repo sits inside a go.work whose other module requires a
            # newer toolchain; the Makefile targets all run with it disabled.
            export GOWORK=off
          '';
        };
      });

      formatter = forAll (pkgs: pkgs.nixfmt-tree);
    };
}
