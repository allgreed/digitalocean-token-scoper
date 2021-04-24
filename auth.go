package main

import (
	"regexp"
)

type AuthorizationRequest struct {
	method string
	path   string
}

type PermissionRule interface {
	can_i(AuthorizationRequest) bool
	is_applicable(AuthorizationRequest) bool
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

type AllowSingleDomainAllRecrodsAllActions struct{ domain string }

func (_ AllowSingleDomainAllRecrodsAllActions) can_i(_ AuthorizationRequest) bool {
	return true
}
func (rule AllowSingleDomainAllRecrodsAllActions) is_applicable(ar AuthorizationRequest) bool {
	rgx := regexp.MustCompile(`\/v2\/domains\/((?:\w+\.)*\w+)\/records(?:\/\d+)?$`)
	rs := rgx.FindStringSubmatch(ar.path)

	if len(rs) == 2 {
		return rs[1] == rule.domain
	} else {
		return false
	}
}

// TODO: add rules for manipulating particular LB
// TODO: create an issue about body authentication -> why I don't need it yet and how to do it => also: take into account changing the logging!
