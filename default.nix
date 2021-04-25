let
  nixpkgs = builtins.fetchGit {
    name = "nixos-unstable-2021-03-17";
    url = "https://github.com/nixos/nixpkgs/";
    ref = "refs/heads/nixos-unstable";
    rev = "266dc8c3d052f549826ba246d06787a219533b8f";
    # obtain via `git ls-remote https://github.com/nixos/nixpkgs nixos-unstable`
  };
  pkgs = import nixpkgs { config = {}; };
in
with pkgs; rec {
  pname = "digitalocean-token-scoper";
  version = "0.2.0";
  executable = buildGoModule rec {
    inherit pname;
    inherit version;

    # TODO: fix leakage to ~/go

    buildInputs = [
      git
      gnumake
      go
      entr
      curl
      jq
    ];
    src = builtins.filterSource (path: type:  baseNameOf path != ".git") ./.;
    vendorSha256 = "0vwbv4q5x2ph7qh63mig9nkk4bz2cmxgqxkvc6c09b3y92cvlknc"; 

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

    contents = [ executable.binary cacert ];

    config = {
      Cmd = [
        "${executable.binary}/bin/${pname}"
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
