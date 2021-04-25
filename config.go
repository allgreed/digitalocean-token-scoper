package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
)

func configure() {
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
	for k := range user_to_permissions {
		users = append(users, k)
	}
	for _, u := range users {
		env_for_user := fmt.Sprintf("APP_USERTOKEN__%s", u)
		default_path := fmt.Sprintf("/secrets/users/%s/secret", u)
		token := read_tokenfile(acquire_env_or_default(env_for_user, default_path))
		token_to_user[token] = u
	}
}

func initialize_logging() {
	log_format := acquire_env_or_default_silent("APP_LOG_FORMAT", "json")

	if log_format == "text" {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	_log_level := acquire_env_or_default("APP_LOG_LEVEL", "info")
	log_level_resolution := map[string]log.Level{
		"debug": log.DebugLevel,
		"info":  log.InfoLevel,
	}

	log_level, found := log_level_resolution[_log_level]
	if !found {
		log_level = log.InfoLevel
		log.WithFields(log.Fields{
			"input_level": _log_level,
		}).Info("Unrecognised log level, falling back to Info")
	}
	log.SetLevel(log_level)

	log.SetOutput(&OutputSplitter{})
}
