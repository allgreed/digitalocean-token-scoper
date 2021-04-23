package main

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

// TODO: test
// TODO: add some real rules, like allow particular LB, allow setting recrods on particular domain
// TODO: create an issue about body authentication -> why I don't need it yet and how to do it => also: take into account changing the logging!
