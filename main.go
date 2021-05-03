package main

import (
	"fmt"

	uuid "github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var (
	target_url          *url.URL
	do_token            string
	port                string
	token_to_user       map[string]string
	user_to_permissions map[string][]PermissionRule
)

func handleFunc(w http.ResponseWriter, r *http.Request) {
	metrics.incoming_requests.Inc()
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
		metrics.unauthorized_requests.Inc()
		JSONError(w, "Please provide a token in the Authorization header", 401)
		logger.Info("Missing token, aborting")
		return
	}
	token := strings.Replace(_token[0], "Bearer ", "", 1)

	user, ok := token_to_user[token]
	if !ok {
		metrics.failed_authentications.Inc()
		JSONError(w, "Please provide a valid token in the Authorization header", 401)
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
		metrics.other_errors.Inc()
		JSONError(w, "Erm... something went wrong.  Expectedly wrong, but still wrong.", 500)
		logger.WithFields(log.Fields{
			"err": err,
		}).Error("Transalting request to interanl representation")
	}
	effectivePermissionRules := append(permissions, DenyAll{})

	logger.WithFields(log.Fields{
		"authorization_request": fmt.Sprintf("%+v", ar),
		"effective_permissions": strings.Replace(fmt.Sprintf("%#v", effectivePermissionRules), "[]main.PermissionRule{", "{", 1),
	}).Debug("")

	for _, rule := range effectivePermissionRules {
		if rule.is_applicable(ar) {
			logger.WithFields(log.Fields{
				"rule": fmt.Sprintf("%#v", rule),
			}).Debug("Matched rule")

			if rule.can_i(ar) {
				break
			}

			metrics.failed_authorizations.Inc()
			JSONError(w, "You don't have access to that resource with that method", 403)
			logger.WithFields(log.Fields{
				"ar": ar,
			}).Warn("Unauthorized action attempt")
			return
		}
	}
	logger.Info("Authorized")

	hh := http.Header{}
	for k, v := range r.Header {
		hh[k] = v
	}

	hh["Authorization"] = []string{"Bearer " + do_token}

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
		metrics.upstream_errors.Inc()
		JSONError(w, "Something wrong with upstream", 504)
		logger.WithFields(log.Fields{
			"err": err,
		}).Warn("Problem reaching target")
		return
	}
	defer resp.Body.Close()

	respH := w.Header()
	for hk, hv := range resp.Header {
		respH[hk] = hv
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		metrics.other_errors.Inc()
		JSONError(w, "Erm... something went wrong.  Expectedly wrong, but still wrong.", 502)
		logger.WithFields(log.Fields{
			"err": err,
		}).Error("Copying response body")
		return
	}
	metrics.successful_requests.Inc()
	logger.Info("Success")
}

func main() {
	configure()

	handler := http.DefaultServeMux
	handler.HandleFunc("/", handleFunc)
	handler.HandleFunc("/healthz", handleOkJSONFunc)
	handler.Handle("/metrics", promhttp.Handler())

	s := &http.Server{
		Addr:           port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.WithFields(log.Fields{
		"port": port,
	}).Info("Starting!")
	log.Fatal(s.ListenAndServe())

	// TODO: all the TODOs from makefiles and default.nix
	// TODO: automated functional tests -> one passes one fails

	// TODO: release 1.0.0
}
