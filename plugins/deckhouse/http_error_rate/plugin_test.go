package http_error_rate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slok/sloth-common-sli-plugins/plugins/deckhouse/http_error_rate"
)

func TestSLIPlugin(t *testing.T) {
	tests := map[string]struct {
		meta     map[string]string
		labels   map[string]string
		options  map[string]string
		expQuery string
		expErr   bool
	}{
		"Without namespace, should fail.": {
			options: map[string]string{"service": "test"},
			expErr:  true,
		},

		"An empty namespace query, should fail.": {
			options: map[string]string{"namespace": "", "service": "test"},
			expErr:  true,
		},

		"Without service, should fail.": {
			options: map[string]string{"namespace": "default"},
			expErr:  true,
		},

		"An empty service query, should fail.": {
			options: map[string]string{"namespace": "default", "service": ""},
			expErr:  true,
		},

		"Not having a filter and with namespace + service should return a valid query.": {
			options: map[string]string{"namespace": "default", "service": "test"},
			expQuery: `
(
  sum(rate(ingress_nginx_detail_responses_total{ namespace="default",service="test",response_code=~"(5..|429)" }[{{.window}}])) 
  /          
  (sum(rate(ingress_nginx_detail_responses_total{ namespace="default",service="test" }[{{.window}}])) > 0)
) OR on() vector(0)
`,
		},

		"Having a filter, with namespace and service should return a valid query.": {
			options: map[string]string{
				"filter":    `k1="v2",k2="v2"`,
				"namespace": "default",
				"service":   "test",
			},
			expQuery: `
(
  sum(rate(ingress_nginx_detail_responses_total{ k1="v2",k2="v2",namespace="default",service="test",response_code=~"(5..|429)" }[{{.window}}])) 
  /          
  (sum(rate(ingress_nginx_detail_responses_total{ k1="v2",k2="v2",namespace="default",service="test" }[{{.window}}])) > 0)
) OR on() vector(0)
`,
		},

		"Filter should be sanitized with ','.": {
			options: map[string]string{
				"filter":    `k1="v2",k2="v2",`,
				"namespace": "default",
				"service":   "test",
			},
			expQuery: `
(
  sum(rate(ingress_nginx_detail_responses_total{ k1="v2",k2="v2",namespace="default",service="test",response_code=~"(5..|429)" }[{{.window}}])) 
  /          
  (sum(rate(ingress_nginx_detail_responses_total{ k1="v2",k2="v2",namespace="default",service="test" }[{{.window}}])) > 0)
) OR on() vector(0)
`,
		},

		"Filter should be sanitized with '{'.": {
			options: map[string]string{
				"filter":    `{k1="v2",k2="v2",},`,
				"namespace": "default",
				"service":   "test",
			},
			expQuery: `
(
  sum(rate(ingress_nginx_detail_responses_total{ k1="v2",k2="v2",namespace="default",service="test",response_code=~"(5..|429)" }[{{.window}}])) 
  /          
  (sum(rate(ingress_nginx_detail_responses_total{ k1="v2",k2="v2",namespace="default",service="test" }[{{.window}}])) > 0)
) OR on() vector(0)
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotQuery, err := http_error_rate.SLIPlugin(context.TODO(), test.meta, test.labels, test.options)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expQuery, gotQuery)
			}
		})
	}
}
