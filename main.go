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
)

var auth = map[string]string{
	"aaaa": "xxxd",
}
var target_url *url.URL
var do_token string

func handleFunc(w http.ResponseWriter, r *http.Request) {
	// TODO: log incoming request properly - everything except the token
	// TODO: give each request some id that'll stick through the logs
	//fmt.Printf("--> %v %v\n", r.Method, r.URL)

	_token := r.Header["Authorization"]
	if len(_token) != 1 {
		http.Error(w, "Please provide a token in the Authorization header", 401)
		// TODO: log - missing or malformed token
		return
	}
	token := _token[0]

	permissions, ok := auth[token]
	if !ok {
		http.Error(w, "Please provide a valid token in the Authorization header", 401)
		// TODO: log - invalid token
		return
	}
	// TODO: log - authenticated as ...

	ar := url_to_auth_request(r.URL, r.Method)
	fmt.Printf("%+v\n", ar)
	fmt.Printf("%+v\n", permissions)
	// TODO: test
	// if fails return 403 with info that request path / method is not allowed
	// TODO: log - unauthorized action attempt

	hh := http.Header{}
	for k, v := range r.Header {
		hh[k] = v
	}

	if _, ok := hh["Authorization"]; ok {
		// TODO: append the real token -> when I've got the token
		hh["Authorization"] = []string{"BLE"}
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
	// TODO: update the README that it's working
	// TODO: build with nix
	// TODO: create minimal container
	// TODO: setup CI

	// TODO: add health and ready url
	// TODO: add metrics
	// TODO: add proper usage, etc to the README
}

func acquire_env_or_default(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

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
func url_to_auth_request(u *url.URL, m string) authorizationRequest {
	// TODO: do some serializaion to make sure we're on the same page
	return authorizationRequest{path: u.Path, method: m}
}

func main() {
	_port := acquire_env_or_fail("APP_PORT")
	port := ":" + _port

	_target_url := acquire_env_or_default("APP_TARGET_URL", "https://api.digitalocean.com/")
	_tu, err := url.Parse(_target_url)
	if err != nil {
		log.Fatal(err)
	}
	target_url = _tu // golang what are u doin' golang stahp!

	// TODO: get the real token from secretpath
	// get the secretpath from the envvar, with a sensible default -> see best practices for mounting secrets in k8s
	do_token = "aaaAAAA"

	// TODO: populate auth from some kind of config? -> map tokens from secret files based on username to permissions

	// TODO: log port, target_url, and the fact that token was loaded

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
}
