package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type OutputSplitter struct{}

func acquire_env_or_default(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	log.Printf("Defaulting to %s=%s", key, fallback)
	return fallback
}

func acquire_env_or_default_silent(key string, fallback string) string {
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

func JSONError(w http.ResponseWriter, err interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}

func url_to_auth_request(u *url.URL, m string) (AuthorizationRequest, error) {
	return AuthorizationRequest{path: u.Path, method: m}, nil
}

func read_file(p string) string {
	_content, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatalf("Something went wrong when reading file from %s, err: %s", p, err)
	}

	return strings.TrimSpace(string(_content))
}

func (splitter *OutputSplitter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("level=error")) || bytes.Contains(p, []byte("level=warn")) || bytes.Contains(p, []byte("level=panic")) || bytes.Contains(p, []byte("level=fatal")) {
		return os.Stderr.Write(p)
	}
	return os.Stdout.Write(p)
}

func get_param(r Rule, p string) string {
	_p := r.Parameters
	parameters := make(map[string]string)

	if _p == nil {
		log.Fatalf("Parameters required, but missing; in rule %s", r.Kind)
	}

	__p, ok := _p.(map[interface{}]interface{})
	if !ok {
		log.Fatalf("Parameters should be a map, but they're not; in rule %s", r.Kind)
	}

	for key, value := range __p {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)

		parameters[strKey] = strValue
	}

	result, found := parameters[p]
	if !found {
		log.Fatalf("Parameter %q required, but not found in rule %q", p, r.Kind)
	}

	return result
}
