package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func acquire_env_or_default(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	log.Printf("Defaulting to %s=%s", key, fallback)
	return fallback
}

func acquire_env_or_fail(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	log.Fatalf("environment variable %q required, but missing", key)
	return "" // will never be reached, but compiler requires it...
}

func JSONError(w http.ResponseWriter, err interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}

func url_to_auth_request(u *url.URL, m string) (AuthorizationRequest, error) {
	// TODO: do some serializaion to make sure we're on the same page
	// TODO: rewrite tests to use new values?
	// TODO: is this actuall needed?
	return AuthorizationRequest{path: u.Path, method: m}, nil
}

func read_tokenfile(p string) string {
	_content, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatalf("Something went wrong when reading secret from %s, err: %s", p, err)
	}

	return strings.TrimSpace(string(_content))
}

func initialize_logging() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	// TODO: logs as json? will it work with loki and grafana?
	// look at golang logging in more depth
	// grep for `log` usage
	// TODO: output errors to stderr, and the rest to stdout
	//log.SetOutput(os.Stdout)
	// TODO: logging based on envvar
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	//https://github.com/sirupsen/logrus
}
