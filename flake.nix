{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "nixpkgs/266dc8c3d052f549826ba246d06787a219533b8f";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages = rec {
          hello = with pkgs; buildGoModule rec {
            pname = "digitalocean-token-scoper";
            version = "0.5.1";

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
          default = hello;
        };

        apps = rec {
          hello = flake-utils.lib.mkApp { drv = self.packages.${system}.hello; };
          default = hello;
        };
      }
    );
}
