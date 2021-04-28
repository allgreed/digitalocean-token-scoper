package main

import (
	"testing"
)

var rulestest = []struct {
	ar            AuthorizationRequest
	rule          Rule
	is_applicable bool
	can_i         bool
}{
	{
		AuthorizationRequest{
			path:   "/v2/domains/example.com/records",
			method: "WHATEVER",
		},
        Rule{Kind:"AllowSingleDomainAllRecordsAllActions", Parameters:map[interface {}]interface {}{"domain":"example.com"}},
		true,
		true,
	},
    {
        AuthorizationRequest{
            path:   "/v2/domains/example.com/records",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleDomainAllRecordsAllActions", Parameters:map[interface {}]interface {}{"domain":"foo.bar"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/domains/something.example.com/records",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleDomainAllRecordsAllActions", Parameters:map[interface {}]interface {}{"domain":"something.example.com"}},
        true,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/domains/example.com/records",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleDomainAllRecordsAllActions", Parameters:map[interface {}]interface {}{"domain":"something.example.com"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/domains/something.example.com/records",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleDomainAllRecordsAllActions", Parameters:map[interface {}]interface {}{"domain":"example.com"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/domains/example.com/records/3352896",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleDomainAllRecordsAllActions", Parameters:map[interface {}]interface {}{"domain":"example.com"}},
        true,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/domains/example.com/records/3352896//?someonetryingtobeclever%$#^&",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleDomainAllRecordsAllActions", Parameters:map[interface {}]interface {}{"domain":"example.com"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/domains/not.allowed.example.com//?/v2/domains/allowed.example.com/",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleDomainAllRecordsAllActions", Parameters:map[interface {}]interface {}{"domain":"example.com"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers/054c04f8-0b7a-40e6-a6ce-bf15b798b2ea/forwarding_rule",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"054c04f8-0b7a-40e6-a6ce-bf15b798b2ea"}},
        true,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers/054c04f8-0b7a-40e6-a6ce-bf15b798b2ea/forwarding_rule/",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"054c04f8-0b7a-40e6-a6ce-bf15b798b2ea"}},
        true,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers/054c04f8-0b7a-40e6-a6ce-bf15b798b2ea/forwarding_rule/nonfunctional",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"054c04f8-0b7a-40e6-a6ce-bf15b798b2ea"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"054c04f8-0b7a-40e6-a6ce-bf15b798b2ea"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers/",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"054c04f8-0b7a-40e6-a6ce-bf15b798b2ea"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers/054c04f8-0b7a-40e6-a6ce-bf15b798b2ea",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"054c04f8-0b7a-40e6-a6ce-bf15b798b2ea"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers/054c04f8-0b7a-40e6-a6ce-bf15b798b2ea/droplets",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"054c04f8-0b7a-40e6-a6ce-bf15b798b2ea"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers/054c04f8-0b7a-40e6-a6ce-bf15b798b2ea/forwarding_rule",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"9e143319-4c08-4d00-9ce5-8e2fd38d90ea"}},
        false,
        true,
    },
    {
        AuthorizationRequest{
            path:   "/v2/load_balancers/9e143319-4c08-4d00-9ce5-8e2fd38d90ea/forwarding_rule?/v2/load_balancers/054c04f8-0b7a-40e6-a6ce-bf15b798b2ea/forwarding_rule",
            method: "WHATEVER",
        },
        Rule{Kind:"AllowSingleLoadBalancerAllForwardingRulesAllActions", Parameters:map[interface {}]interface {}{"load_balancer_id":"054c04f8-0b7a-40e6-a6ce-bf15b798b2ea"}},
        false,
        true,
    },
}

func TestRules(t *testing.T) {
	for _, test_case := range rulestest {
        rule, err := parse_rule(test_case.rule)
		if err != nil {
			t.Errorf("rule = %+v cannot be parsed", test_case.rule)
		}

		can_i := rule.can_i(test_case.ar)
		is_applicable := rule.is_applicable(test_case.ar)

		if test_case.is_applicable != is_applicable {
			t.Errorf("ar = %+v, rule = %#v, is_applicable = %t; wanted opposite", test_case.ar, test_case.rule, is_applicable)
		}
		if is_applicable && (test_case.can_i != can_i) {
			t.Errorf("ar = %+v, rule = %#v, can_i = %t; wanted opposite", test_case.ar, test_case.rule, can_i)
		}
	}

	return
}
