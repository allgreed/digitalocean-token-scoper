package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"

	"gopkg.in/yaml.v2"
)

type ConfigPayload struct {
	Permissions []BoundRule `yaml:"permissions"`
}
type BoundRule struct {
	User  string `yaml:"user"`
	Rules []Rule `yaml:"rules"`
}
type Rule struct {
	Kind       string      `yaml:"rule"`
	Parameters interface{} `yaml:"parameters"`
}

func configure() {
	initialize_logging() // this has to be first to ensure log consistency

	_port := acquire_env_or_default("APP_PORT", "80")
	port = ":" + _port

	_target_url := acquire_env_or_default("APP_TARGET_URL", "https://api.digitalocean.com/")
	_tu, err := url.Parse(_target_url)
	if err != nil {
		log.WithFields(log.Fields{
			"app_target_url": _target_url,
			"err":            err,
		}).Fatal("Processing APP_TARGET_UR")
	}
	target_url = _tu // golang what are u doin' golang stahp!

	do_token = read_file(acquire_env_or_default("APP_TOKEN_PATH", "/secrets/token/secret"))

	_user_to_permissions := read_file(acquire_env_or_default("APP_PERMISSIONS_PATH", "/config/permissions/config"))
	var __user_to_permissions ConfigPayload
	err = yaml.UnmarshalStrict([]byte(_user_to_permissions), &__user_to_permissions)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Parsing permission yaml")
	}
	user_to_permissions = parse_config(__user_to_permissions)

	users := []string{}
	for k := range user_to_permissions {
		users = append(users, k)
	}
	token_to_user = make(map[string]string)
	for _, u := range users {
		env_for_user := fmt.Sprintf("APP_USERTOKEN__%s", u)
		default_path := fmt.Sprintf("/secrets/users/%s/secret", u)
		token := read_file(acquire_env_or_default(env_for_user, default_path))
		token_to_user[token] = u
	}
}

func initialize_logging() {
	log_format := acquire_env_or_default_silent("APP_LOG_FORMAT", "json")

	if log_format == "text" {
		log.SetFormatter(&log.TextFormatter{
			DisableColors:    true,
			DisableTimestamp: true,
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

func parse_config(c ConfigPayload) map[string][]PermissionRule {
	log.WithFields(log.Fields{
		"payload": c,
	}).Debug("Begin processing")
	result := make(map[string][]PermissionRule)
	for _, bound_rules := range c.Permissions {
		u := bound_rules.User
		rules := []PermissionRule{}
		for _, rule := range bound_rules.Rules {
			log.WithFields(log.Fields{
				"rule": rule,
				"user": u,
			}).Debug("Processing rule")
			r, err := parse_rule(rule)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Fatal("Parsing rule")
			}
			rules = append(rules, r)
		}
		result[u] = rules
	}

	log.WithFields(log.Fields{
		"permissions": fmt.Sprintf("%+v", result),
	}).Info("Permissions sucesfully parsed")
	return result
}
