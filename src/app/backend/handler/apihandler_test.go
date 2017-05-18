package handler

import (
	"net/http"
	"testing"

	"bytes"
	"github.com/emicklei/go-restful"
	"reflect"
	"strings"
)

func TestCreateHTTPAPIHandler(t *testing.T) {
	_, err := CreateHTTPAPIHandler(nil, ApiClientConfig{ApiserverHost: "127.0.0.1", KubeConfigFile: ""})
	if err != nil {
		t.Fatal("CreateHTTPAPIHandler() cannot create HTTP API handler")
	}
}

func TestShouldDoCsrfValidation(t *testing.T) {
	cases := []struct {
		request  *restful.Request
		expected bool
	}{
		{
			&restful.Request{
				Request: &http.Request{
					Method: "PUT",
				},
			},
			false,
		},
		{
			&restful.Request{
				Request: &http.Request{
					Method: "POST",
				},
			},
			true,
		},
	}
	for _, c := range cases {
		actual := shouldDoCsrfValidation(c.request)
		if actual != c.expected {
			t.Errorf("shouldDoCsrfValidation(%#v) returns %#v, expected %#v", c.request, actual, c.expected)
		}
	}
}

func TestMapUrlToResource(t *testing.T) {
	cases := []struct {
		url, expected string
	}{
		{
			"/api/v1/pod",
			"pod",
		},
		{
			"/api/v1/node",
			"node",
		},
	}
	for _, c := range cases {
		actual := mapUrlToResource(c.url)
		if !reflect.DeepEqual(actual, &c.expected) {
			t.Errorf("mapUrlToResource(%#v) returns %#v, expected %#v", c.url, actual, c.expected)
		}
	}
}

func TestFormatRequestLog(t *testing.T) {
	req, err := http.NewRequest("PUT", "/api/v1/pod", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Error("Cannot mockup request")
	}
	cases := []struct {
		request  *restful.Request
		expected string
	}{
		{
			&restful.Request{
				Request: req,
			},
			"Incoming HTTP/1.1 PUT /api/v1/pod request",
		},
	}
	for _, c := range cases {
		actual := formatRequestLog(c.request)
		if !strings.Contains(actual, c.expected) {
			t.Errorf("formatRequestLog(%#v) returns %#v, expected to contain %#v", c.request, actual, c.expected)
		}
	}
}
