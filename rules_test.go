package main

import (
	"testing"
)

var rulestest = []struct {
	ar            AuthorizationRequest
	rule          PermissionRule
	is_applicable bool
	can_i         bool
}{
	{
		AuthorizationRequest{
			path:   "/v2/domains/example.com/records",
			method: "WHATEVER",
		},
		AllowSingleDomainAllRecrodsAllActions{domain: "example.com"},
		true,
		true,
	},
	{
		AuthorizationRequest{
			path:   "/v2/domains/example.com/records",
			method: "WHATEVER",
		},
		AllowSingleDomainAllRecrodsAllActions{domain: "foo.bar"},
		false,
		true,
	},
	{
		AuthorizationRequest{
			path:   "/v2/domains/something.example.com/records",
			method: "WHATEVER",
		},
		AllowSingleDomainAllRecrodsAllActions{domain: "something.example.com"},
		true,
		true,
	},
	{
		AuthorizationRequest{
			path:   "/v2/domains/example.com/records",
			method: "WHATEVER",
		},
		AllowSingleDomainAllRecrodsAllActions{domain: "something.example.com"},
		false,
		true,
	},
	{
		AuthorizationRequest{
			path:   "/v2/domains/something.example.com/records",
			method: "WHATEVER",
		},
		AllowSingleDomainAllRecrodsAllActions{domain: "example.com"},
		false,
		true,
	},
	{
		AuthorizationRequest{
			path:   "/v2/domains/example.com/records/3352896",
			method: "WHATEVER",
		},
		AllowSingleDomainAllRecrodsAllActions{domain: "example.com"},
		true,
		true,
	},
	{
		AuthorizationRequest{
			path:   "/v2/domains/example.com/records/3352896//?someonetryingtobeclever%$#^&",
			method: "WHATEVER",
		},
		AllowSingleDomainAllRecrodsAllActions{domain: "example.com"},
		false,
		true,
	},
}

func TestRules(t *testing.T) {
	for _, test_case := range rulestest {
		can_i := test_case.rule.can_i(test_case.ar)
		is_applicable := test_case.rule.is_applicable(test_case.ar)

		if test_case.is_applicable != is_applicable {
			t.Errorf("ar = %+v, rule = %#v, is_applicable = %t; wanted opposite", test_case.ar, test_case.rule, is_applicable)
		}
		if is_applicable && (test_case.can_i != can_i) {
			t.Errorf("ar = %+v, rule = %#v, can_i = %t; wanted opposite", test_case.ar, test_case.rule, can_i)
		}
	}

	return
}
