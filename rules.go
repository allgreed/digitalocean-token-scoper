package main

import (
	"regexp"
    "errors"
)

type AuthorizationRequest struct {
	method string
	path   string
}

type PermissionRule interface {
	can_i(AuthorizationRequest) bool
	is_applicable(AuthorizationRequest) bool
}


//TODO: test for exhaustivness
//https://www.reddit.com/r/golang/comments/8nz2mc/finding_all_types_which_implement_an_interface_at/
func parse_rule(r Rule) (PermissionRule, error) {
    var result PermissionRule
    var err error = nil
	switch kind := r.Kind; kind {
	case "AllowSingleDomainAllRecordsAllActions":
        domain := get_param(r, "domain")
        result = AllowSingleDomainAllRecordsAllActions{domain}
    case "AllowSingleLoadBalancerAllForwardingRulesAllActions":
        load_balancer_id := get_param(r, "load_balancer_id")
        result = AllowSingleLoadBalancerAllForwardingRulesAllActions{load_balancer_id}
	default:
        err = errors.New("Unkown rule, aborting!")
	}

    return result, err
}

type AllowAll struct{}

func (_ AllowAll) can_i(_ AuthorizationRequest) bool {
	return true
}
func (_ AllowAll) is_applicable(_ AuthorizationRequest) bool {
	return true
}

type DenyAll struct{}

func (_ DenyAll) can_i(_ AuthorizationRequest) bool {
	return false
}
func (_ DenyAll) is_applicable(_ AuthorizationRequest) bool {
	return true
}

type AllowSingleDomainAllRecordsAllActions struct{ domain string }

func (_ AllowSingleDomainAllRecordsAllActions) can_i(_ AuthorizationRequest) bool {
	return true
}
func (rule AllowSingleDomainAllRecordsAllActions) is_applicable(ar AuthorizationRequest) bool {
	rgx := regexp.MustCompile(`^\/v2\/domains\/((?:\w+\.)*\w+)\/records(?:\/\d+)?$`)
	rs := rgx.FindStringSubmatch(ar.path)

	if len(rs) == 2 {
		return rs[1] == rule.domain
	} else {
		return false
	}
}

type AllowSingleLoadBalancerAllForwardingRulesAllActions struct{ load_balancer_id string }

func (_ AllowSingleLoadBalancerAllForwardingRulesAllActions) can_i(_ AuthorizationRequest) bool {
	return true
}
func (rule AllowSingleLoadBalancerAllForwardingRulesAllActions) is_applicable(ar AuthorizationRequest) bool {
	rgx := regexp.MustCompile(`^\/v2\/load_balancers\/([a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12})\/forwarding_rule\/?$`)
	rs := rgx.FindStringSubmatch(ar.path)

	if len(rs) == 2 {
		return rs[1] == rule.load_balancer_id
	} else {
		return false
	}
}

// TODO: create an issue about body authentication -> why I don't need it yet and how to do it => also: take into account changing the logging!
