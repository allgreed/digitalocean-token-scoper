// heavily based on https://github.com/davidfstr/nanoproxy
package main

import (
	uuid "github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
    "fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var (
	target_url *url.URL
	do_token   string
	port       string
    token_to_user = map[string]string {
        "aaaa": "testowy",
    }
	user_to_permissions = map[string][]PermissionRule{
		"testowy": {AllowAll{}},
	}
)

func handleFunc(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.NewRandom()
	logger := log.WithFields(log.Fields{
		"request_id": requestID,
	})
	logger.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.URL,
	}).Info("Incoming request")

	_token := r.Header["Authorization"]
	if len(_token) != 1 {
		http.Error(w, "Please provide a token in the Authorization header", 401)
		logger.Info("Missing token, aborting")
		return
	}
	token := _token[0]


	user, ok := token_to_user[token]
	if !ok {
		http.Error(w, "Please provide a valid token in the Authorization header", 401)
		logger.Info("Invalid or unknown token, aborting")
		return
	}
	logger.WithFields(log.Fields{
		"user": user,
	}).Info("Authenticated")


	permissions, ok := user_to_permissions[user]
	if !ok {
        logger.Panic("Entry for authenticated user missing from permisson table, something went terribly wrong!")
	}

	ar, err := url_to_auth_request(r.URL, r.Method)
    if err != nil {
        http.Error(w, "You don't have access to that resource with that method", 403)
        logger.WithFields(log.Fields{
            "ar": ar,
        }).Warn("Unauthorized action attempt")
        return
    }
	effectivePermissionRules := append(permissions, DenyAll{})

	logger.WithFields(log.Fields{
		"authorization_request": fmt.Sprintf("%+v", ar),
		"effective_permissions": strings.Replace(fmt.Sprintf("%#v", effectivePermissionRules), "[]main.PermissionRule{", "{", 1),
	}).Debug("Will auth in a tick")

	for _, rule := range effectivePermissionRules {
		if rule.is_applicable(ar) {
			logger.WithFields(log.Fields{
				"rule": fmt.Sprintf("%#v", rule),
			}).Debug("Matched rule")

			if !rule.can_i(ar) {
				http.Error(w, "You don't have access to that resource with that method", 403)
				logger.WithFields(log.Fields{
					"ar": ar,
				}).Warn("Unauthorized action attempt")
				return
			} else {
				break
			}
		}
	}
	logger.Info("Authorized")


	hh := http.Header{}
	for k, v := range r.Header {
		hh[k] = v
	}

	if _, ok := hh["Authorization"]; ok {
		hh["Authorization"] = []string{do_token}
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
		// TODO: fix this - what info should be passed? maybe even error!
		logger.WithFields(log.Fields{}).Warn("Wabababababa")
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
	logger.Info("Success")
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

func url_to_auth_request(u *url.URL, m string) (AuthorizationRequest, error) {
	// TODO: do some serializaion to make sure we're on the same page
	return AuthorizationRequest{path: u.Path, method: m}, nil
}

func init() {
	_port := acquire_env_or_fail("APP_PORT")
	port = ":" + _port

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
	// TODO: output errors to stderr, and the rest to stdout
	//log.SetOutput(os.Stdout)
	// TODO: logging based on envvar
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
    //https://github.com/sirupsen/logrus

	// TODO: populate auth from some kind of config? -> map tokens from secret files based on username to permissions
    // TODO: verify that all the tokens have corresponding user entries

}

func main() {
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
	// TODO: add 'built with nix' badge
	// TODO: create minimal container
	// TODO: setup CI

	// TODO: add health and ready url
	// TODO: add metrics
	// TODO: add proper usage, etc to the README
}
