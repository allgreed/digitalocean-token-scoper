# digitalocean-token-scoper
A solution to [Digitalocean](https://www.digitalocean.com/)'s [lack of token scoping](https://ideas.digitalocean.com/ideas/DO-I-966)*
<!--*-->
[![built with nix](https://builtwithnix.org/badge.svg)](https://builtwithnix.org)
[![Build Status](https://cloud.drone.io/api/badges/allgreed/digitalocean-token-scoper/status.svg)](https://cloud.drone.io/allgreed/digitalocean-token-scoper)

\* technically they have scoping, you can choose either write or read + write. Plz no sue.

## Usage
It's heavily alpha right now - it's running in production, but ~may~ will require editing source code for setting permissions

Usernames are fairly arbitrary, however for the development purpose I'm assuming alphanumeric ASCII and that tokens don't contain significant whitespace.

For now go for [dev](#dev)

### Permission model
- rules are applied sequentially in order
- if a rule applies then it's the authority on weather grant or deny access
- by default DenyAll is appended at the end of the rule chain


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

### Adding new rules
- in `./rules_test.go` append your test cases to `rulestest` (at the end)
- in `./rules.go`, add and fill: 
```
type X struct{ // rule parameters }

func (rule X) can_i(ar AuthorizationRequest) bool {
    // write your authorization logic here
}
func (rule X) is_applicable(ar AuthorizationRequest) bool {
    // write your applicability logic here
}
```

## Security considerations

- I strongly suggest not exposing this service to the internet
- standard considerations apply in order to secure the token storage and access to the app environment
- token verification should be resistant against time-attacks, however this wasn't tested
- there is no rate-limiting mechanism on a per-user basis => the DO account's limit is shared by all the users
- response from DO's API is passed to the client **as is**, including headers. I've seen nothing sensitive there (as of 24.04.2021), yet afaik it's not guaranteed by DO.
