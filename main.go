package main

// heavily based on https://github.com/davidfstr/nanoproxy

import (
	"fmt"

	uuid "github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"

	//"io"
	"io/ioutil"
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
		JSONError(w, "Please provide a token in the Authorization header", 401)
		logger.Info("Missing token, aborting")
		return
	}
	token := strings.Replace(_token[0], "Bearer ", "", 1)

	user, ok := token_to_user[token]
	if !ok {
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
		// TODO: fix this up
		//JSONError(w, "You don't have access to that resource with that method", 403)
		//logger.WithFields(log.Fields{
		//"ar": ar,
		//}).Warn("Unauthorized action attempt")
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
				JSONError(w, "You don't have access to that resource with that method", 403)
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
		hh["Authorization"] = []string{"Bearer " + do_token}
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
		JSONError(w, "Could not reach origin server", 500)
		// TODO: fix this - what info should be passed? maybe even error!
		logger.WithFields(log.Fields{
			"err": err,
		}).Warn("Wabababababa")
		return
	}
	defer resp.Body.Close()

	respH := w.Header()
	for hk, hv := range resp.Header {
		respH[hk] = hv
	}
	w.WriteHeader(resp.StatusCode)
	//if resp.ContentLength > 0 {
	//// ignore I/O errors, since there's nothing we can do
	//io.CopyN(w, resp.Body, resp.ContentLength)
	//} else if resp.Close {
	//for {
	//if _, err := io.Copy(w, resp.Body); err != nil {
	//fmt.Println(err)
	//break
	//}
	//}
	//}
	// TODO: stream instead of buffering to memory
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// TODO: handle this properly
		log.Fatal(err)
	}
	w.Write(bodyBytes)
	logger.Info("Success")
}

func main() {
	configure()

	handler := http.DefaultServeMux
	handler.HandleFunc("/", handleFunc)
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

	// TODO: fill Docker and k8s examples
    // TODO: - do-token-scoper => errs as fields when logging
    // TODO: - when "kind" instead of "rule" => make the error more visible
    // TODO: release 0.4.2

	// TODO: cover minor todos - `make todo`
	// TODO: remove beta disclaimer

	// TODO: ask for a 3rd party security audit
	// TODO: release 1.0.0

	// TODO: automated E2E tests (k8s example) - for manual running?
	// TODO: add health and ready url
	// TODO: add metrics
	// TODO: add proper usage (local / k8s examples), etc to the README
	// TODO: add badges -> snyk vulnearbilities
	// TODO: setup dependency monitoring
	// https://github.com/dwyl/repo-badges
}
