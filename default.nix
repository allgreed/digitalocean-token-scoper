let
  nixpkgs = builtins.fetchGit {
    url = "https://github.com/nixos/nixpkgs/";
    ref = "refs/heads/nixos-unstable";
    rev = "266dc8c3d052f549826ba246d06787a219533b8f";
    # obtain via `git ls-remote https://github.com/nixos/nixpkgs nixos-unstable`
  };
  pkgs = import nixpkgs { config = {}; };
in
with pkgs; rec {
  pname = "digitalocean-token-scoper";
  version = "0.5.1";
  app = buildGoModule rec {
    inherit pname;
    inherit version;

    # TODO: fix leakage to ~/go
    # TODO: fix not having this stuff avaible in dev-shell (and in CI)
    buildInputs = [
      git
      gnumake
      go
      entr
      curl
      jq
    ];
    src = builtins.filterSource (path: type:  baseNameOf path != ".git") ./.;
    vendorSha256 = "14j9l9g6zk3rjqw3iwmpjxhzhiqi7sfrq0415hrcylypdxiyknw3";

    subPackages = [ "." ]; 

    runVend = true; 

    meta = with lib; {
      description = "A solution to Digitalocean's lack of token scoping*";
      homepage = "https://github.com/allgreed/digitalocean-token-scoper";
      license = licenses.mit;
      maintainers = with maintainers; [ allgreed ];
      platforms = platforms.linux;
    };
  };
  docker.image = pkgs.dockerTools.buildLayeredImage {
    name = pname;
    tag = version;
    maxLayers = 30; # https://nixos.org/manual/nixpkgs/stable/#ssec-pkgs-dockerTools-buildLayeredImage

    created = "now";

    contents = [ app cacert ];

    config = {
      Cmd = [
        "${app}/bin/${pname}"
      ];

      ExposedPorts = {
        "80/tcp" = {};
      };

      Env = [
        "SSL_CERT_FILE=${cacert}/etc/ssl/certs/ca-bundle.crt"
      ];
    };
  };
}
