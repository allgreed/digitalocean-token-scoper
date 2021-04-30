package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	incoming_requests      prometheus.Counter
	failed_authentications prometheus.Counter
	failed_authorizations  prometheus.Counter
	unauthorized_requests  prometheus.Counter
	upstream_errors        prometheus.Counter
	successful_requests    prometheus.Counter
	other_errors           prometheus.Counter
}

var (
	metrics = Metrics{
		promauto.NewCounter(prometheus.CounterOpts{
			Name: "digitalocean_token_scoper_requests_total",
			Help: "The total number of incoming requests",
		}),
		promauto.NewCounter(prometheus.CounterOpts{
			Name: "digitalocean_token_scoper_failed_authentications_total",
			Help: "The total number of failed authentications - user had a token, but it was invalid",
		}),
		promauto.NewCounter(prometheus.CounterOpts{
			Name: "digitalocean_token_scoper_failed_authorizations_total",
			Help: "The total number of failed authorization - user was authenticated, but had no permission to perform desired operation",
		}),
		promauto.NewCounter(prometheus.CounterOpts{
			Name: "digitalocean_token_scoper_unauthorized_requests_total",
			Help: "The total number of requests that don't have token provided",
		}),
		promauto.NewCounter(prometheus.CounterOpts{
			Name: "digitalocean_token_scoper_upstream_errors_total",
			Help: "The total number of errors on behalf of the Digitalocean's API",
		}),
		promauto.NewCounter(prometheus.CounterOpts{
			Name: "digitalocean_token_scoper_successful_requests_total",
			Help: "The total number of successfully processed requests",
		}),
		promauto.NewCounter(prometheus.CounterOpts{
			Name: "digitalocean_token_scoper_other_errors_total",
			Help: "The total number of other errors",
		}),
	}
)
