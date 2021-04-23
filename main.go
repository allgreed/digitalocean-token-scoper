// heavily based on https://github.com/davidfstr/nanoproxy
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
    "io/ioutil"
    "strings"
)

var auth = map[string][]PermissionRule{
	"aaaa": {AllowAll{}},
}
var target_url *url.URL
var do_token string

func handleFunc(w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming request to %v %v\n", r.Method, r.URL)
	// TODO: give each request some id that'll stick through the logs

	_token := r.Header["Authorization"]
	if len(_token) != 1 {
		http.Error(w, "Please provide a token in the Authorization header", 401)
		// TODO: log - missing or malformed token
		// TODO: stick request context
		return
	}
	token := _token[0]

	permissions, ok := auth[token]
	if !ok {
		http.Error(w, "Please provide a valid token in the Authorization header", 401)
		// TODO: log - invalid token
		// TODO: stick request context
		return
	}
	// TODO: log - authenticated as ...
    // TODO: stick request context

	ar := url_to_auth_request(r.URL, r.Method)

    effectivePermissionRules := append(permissions, DenyAll{})

    // TODO: log debug
	fmt.Printf("%+v\n", ar)
	fmt.Printf("%+v\n", permissions)
	fmt.Printf("%+v\n", effectivePermissionRules)
    for _, rule := range effectivePermissionRules {
        if rule.is_applicable(ar) {
            // TODO: log debug
	        fmt.Printf("matched rule %T with parameters %+v\n", rule, rule)
            if !rule.can_i(ar) {
		        http.Error(w, "You don't have access to that resource with that method", 403)
                // TODO: log - unauthorized action attempt
                return
            } else {
                break
            }
        }
    }
	// TODO: log - authrized access
    // TODO: stick request context


	hh := http.Header{}
	for k, v := range r.Header {
		hh[k] = v
	}

	if _, ok := hh["Authorization"]; ok {
		hh["Authorization"] = []string{do_token}
        // TODO: make sure it's there and instead of the original auth
	}

	r.URL.Host = target_url.Host
	r.URL.Scheme = target_url.Scheme
	r.URL.Path = path.Join(target_url.Path, r.URL.Path)
	proxied_request := http.Request{
		Method:        r.Method,
		URL:           r.URL,
		Header:        hh,
		Body:          r.Body,
		ContentLength: r.ContentLength,
		Close:         r.Close,
	}
	resp, err := http.DefaultTransport.RoundTrip(&proxied_request)
	if err != nil {
		// TODO: relay more info ?
		http.Error(w, "Could not reach origin server", 500)
		// TODO: log - request failed
        // TODO: stick request context
		return
	}
	defer resp.Body.Close()

	respH := w.Header()
	for hk, hv := range resp.Header {
		respH[hk] = hv
	}
	w.WriteHeader(resp.StatusCode)
	if resp.ContentLength > 0 {
		// ignore I/O errors, since there's nothing we can do
		io.CopyN(w, resp.Body, resp.ContentLength)
	} else if resp.Close {
		for {
			if _, err := io.Copy(w, resp.Body); err != nil {
				break
			}
		}
	}
	// TODO: log - request completed succesfully
    // TODO: stick request context
    // TODO: init module
    // TODO: gofmt
}

func acquire_env_or_default(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

    log.Printf("Defaulting to %s=%s\n", key, fallback)
	return fallback
}

func acquire_env_or_fail(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	log.Fatalf("environment variable %q required, but missing", key)
	return "" // will never be reached, but compiler requires it...
}

// TODO: can fail?
func url_to_auth_request(u *url.URL, m string) AuthorizationRequest {
	// TODO: do some serializaion to make sure we're on the same page
	return AuthorizationRequest{path: u.Path, method: m}
}

func main() {
	_port := acquire_env_or_fail("APP_PORT")
	port := ":" + _port

	_target_url := acquire_env_or_default("APP_TARGET_URL", "https://api.digitalocean.com/")
	_tu, err := url.Parse(_target_url)
	if err != nil {
        log.Fatalf("Something went wrong when processing APP_TARGET_URL, err: %s", err)
	}
	target_url = _tu // golang what are u doin' golang stahp!

	token_path := acquire_env_or_default("APP_TOKEN_PATH", "/secrets/token")
    _do_token, err := ioutil.ReadFile(token_path)
    if err != nil {
        log.Fatalf("Something went wrong when reading secret from %s, err: %s", token_path, err)
    }
	do_token = strings.TrimSpace(string(_do_token))
    // TODO: logs as json? will it work with loki and grafana?
        // look at golang logging in more depth
        // grep for `log` usage

	// TODO: populate auth from some kind of config? -> map tokens from secret files based on username to permissions

	handler := http.DefaultServeMux
	handler.HandleFunc("/", handleFunc)
	s := &http.Server{
		Addr:           port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()

    // TODO: test it E2E
	// TODO: update the README that it's working
    // TODO: add a section on how to use and create rules
	// TODO: build with nix
	// TODO: create minimal container
	// TODO: setup CI

	// TODO: add health and ready url
	// TODO: add metrics
	// TODO: add proper usage, etc to the README
}
