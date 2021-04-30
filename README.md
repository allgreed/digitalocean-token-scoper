# digitalocean-token-scoper
A solution to [Digitalocean](https://www.digitalocean.com/)'s [lack of token scoping](https://ideas.digitalocean.com/ideas/DO-I-966)*
<!--*-->
[![built with nix](https://img.shields.io/badge/-%20nix-%235277E3?logo=NixOs&label=built%20with)](https://builtwithnix.org)
[![Build Status](https://cloud.drone.io/api/badges/allgreed/digitalocean-token-scoper/status.svg)](https://cloud.drone.io/allgreed/digitalocean-token-scoper)
![Docker Image Version (latest semver)](https://img.shields.io/docker/v/allgreed/digitalocean-token-scoper?sort=semver)

\* technically they have scoping, you can choose either write or read + write. Plz no sue.

## Usage
- It's an HTTP proxy, just run it (either as binary or container) and send DigitalOcean requests to it 
- Usernames are fairly arbitrary, however I'm assuming alphanumeric ASCII and that tokens should not contain significant whitespace at the begining or end.
- See examples

### Example: Docker
```bash
mkdir /tmp/do-ts-example
cat << EOF > /tmp/do-ts-example/config
permissions:
  - user: joe
    rules:
      - rule: AllowSingleDomainAllRecordsAllActions
        parameters:
          domain: example.com
EOF
cat << EOF > /tmp/do-ts-example/joe-secret
aaaa
EOF
cat << EOF > /tmp/do-ts-example/do-token
# insert you Digitalocean token here
EOF
docker run --rm -v /tmp/do-ts-example:/data \
    -e APP_PERMISSIONS_PATH=/data/config \
    -e APP_USERTOKEN__joe=/data/joe-secret \
    -e APP_TOKEN_PATH=/data/do-token \
    -p 7777:80 \
    allgreed/digitalocean-token-scoper

# in a new shell
# next one will pass
curl http://localhost:7777/v2/domains/example.com/records --silent -H "Authorization: Bearer aaaa" | jq

# those 3 will not (2 x unathorized, 1 x unathenticated)
curl http://localhost:7777/v2/domains/example.org/records --silent -H "Authorization: Bearer aaaa" | jq
curl http://localhost:7777/v2/droplets --silent -H "Authorization: aaaa" | jq
# yeah, the "Bearer" part is optional
curl http://localhost:7777/v2/droplets --silent -H "Authorization: Bearer bbbb" | jq
```

### Example: k8s
```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: digitalocean-token-scoper
spec:
  replicas: 1
  selector:
    matchLabels:
      name: digitalocean-token-scoper
  template:
    spec:
      containers:
      - name: main
        image: docker.io/allgreed/digitalocean-token-scoper
        ports:
        - containerPort: 80
        startupProbe:
          httpGet:
            path: /healthz
            port: 80
          failureThreshold: 10
          periodSeconds: 1
        volumeMounts:
        - name: joe-token
          mountPath: "/secrets/users/joe"
          readOnly: true
        - name: do-api-token
          mountPath: "/secrets/token"
          readOnly: true
        - name: permissions
          mountPath: "/config/permissions"
          readOnly: true
      volumes:
      - name: permissions
        configMap:
          name: digitalocean-token-scoper
      - name: joe-token
        secret:
          secretName: joe-token
      - name: do-api-token
        secret:
          secretName: do-api-token
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: digitalocean-token-scoper
data:
  config: |
    permissions:
      - user: joe
        rules:
          - rule: AllowSingleDomainAllRecordsAllActions
            parameters:
              domain: example.com
---
apiVersion: v1
kind: Secret
metadata:
  name: joe-token
type: Opaque
data:
  secret: aaaa # <- please don't use this in production
---
apiVersion: v1
kind: Secret
metadata:
  name: do-api-token
type: Opaque
data:
  secret: ... # your Digitalocean API token goes here
```

### Example: Terraform
- Given an instance of digitalocean-token-scoper is running at example.com:7777, user joe has a token and associated permissions

```terraform
provider digitalocean {
  token        = "joe's token goes here"
  api_endpoint = "http://example.com:7777"
}
```

- run your terraform as you normally would [yup, this was tested and proven to work!] ;)

### Permission model
- rules are applied sequentially in order
- if a rule applies then it's the authority on weather grant or deny access
- by default DenyAll is appended at the end of the rule chain

### Permission format

See `./example.yaml`

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
```go
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
