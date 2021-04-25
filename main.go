// heavily based on https://github.com/davidfstr/nanoproxy
package main

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
	token_to_user       = make(map[string]string)
	user_to_permissions = map[string][]PermissionRule{
		"allgreed": {AllowSingleDomainAllRecrodsAllActions{domain: "olgierd.space"}},
		"dawid": {AllowSingleDomainAllRecrodsAllActions{domain: "tygrys.me"}},
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
		JSONError(w, "Please provide a token in the Authorization header", 401)
		logger.Info("Missing token, aborting")
		return
	}
	token := _token[0]

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
	fmt.Println(resp.Body)
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
	initialize_logging() // this has to be first to ensure log consistency

	_port := acquire_env_or_default("APP_PORT", "80")
	port = ":" + _port

	_target_url := acquire_env_or_default("APP_TARGET_URL", "https://api.digitalocean.com/")
	_tu, err := url.Parse(_target_url)
	if err != nil {
		log.Fatalf("Something went wrong when processing APP_TARGET_URL, err: %s", err)
	}
	target_url = _tu // golang what are u doin' golang stahp!

	do_token = read_tokenfile(acquire_env_or_default("APP_TOKEN_PATH", "/secrets/token/secret"))

	users := []string{}
	for k, _ := range user_to_permissions {
		users = append(users, k)
	}
	for _, u := range users {
		env_for_user := fmt.Sprintf("APP_USERTOKEN__%s", u)
		default_path := fmt.Sprintf("/secrets/users/%s/secret", u)
		token := read_tokenfile(acquire_env_or_default(env_for_user, default_path))
		token_to_user[token] = u
	}

	handler := http.DefaultServeMux
	handler.HandleFunc("/", handleFunc)
	s := &http.Server{
		Addr:           port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())

	// TODO: put this into production
	// TODO: update the README that it's working in production, but there are still paths to be covered

	// TODO: json logs! ^^
	// TODO: setup CI
	// TODO: automated functional tests - run and hit it with a request - 1 fails one passes

	// TODO: add a section on how to use and create rules
	// TODO: write the LB rule

	// TODO: cover minor todos - `make todo`
	// TODO: ask for a 3rd party security audit
	// TODO: update info that it's working , but not yet production quality

	// TODO: automated E2E tests (k8s example)
	// TODO: parametrize user config permissions
	// TODO: add health and ready url
	// TODO: add metrics
	// TODO: add proper usage (local / k8s examples), etc to the README
	// TODO: add badges -> snyk vulnearbilities,  drone passing, test coverage, codeclimate style
	// TODO: setup dependency monitoring
	// https://github.com/dwyl/repo-badges
	// TODO: delete quality disclaimers - this will remain <1.0.0 until I figure out how to sensibly configure permissions without hardcoding them
        // TODO: add a github issue for that!
}
