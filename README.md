# digitalocean-token-scoper
A solution to [Digitalocean](https://www.digitalocean.com/)'s [lack of token scoping](https://ideas.digitalocean.com/ideas/DO-I-966)*
<!--*-->
[![built with nix](https://builtwithnix.org/badge.svg)](https://builtwithnix.org)

\* technically they have scoping, you can choose either write or read + write. Plz no sue.

## Usage
It's heavily alpha right now - it's proven to work, but may require editing source code of PoC quality for your usecase

For now go for [dev](#dev)

## Dev

### Prerequisites
- [nix](https://nixos.org/nix/manual/#chap-installation)
- `direnv` (`nix-env -iA nixpkgs.direnv`)
- [configured direnv shell hook ](https://direnv.net/docs/hook.html)
- some form of `make` (`nix-env -iA nixpkgs.gnumake`)

Hint: if something doesn't work because of missing package please add the package to `default.nix` instead of installing on your computer. Why solve the problem for one if you can solve the problem for all? ;)

### One-time setup
```
make init
```

### Everything
```
make help
```

## Security considerations

TODO
