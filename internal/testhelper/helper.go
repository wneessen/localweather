// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package testhelper

import (
	stdhttp "net/http"
	"os"
	"strings"
	"testing"

	"github.com/wneessen/localweather/internal/config"
	"github.com/wneessen/localweather/internal/log"
)

const (
	TestOnlineAPIURL = "https://api.restful-api.dev/objects"
)

func NewLogger(t *testing.T, level config.LogLevel, out string) *log.Logger {
	t.Helper()

	if out == "" {
		out = "stdout"
	}

	conf := new(config.Config)
	conf.Log.Level = level
	conf.Log.Output = out
	return log.New(conf)
}

func PerformIntegrationTests(t *testing.T) {
	t.Helper()
	if val := os.Getenv("PERFORM_INTEGRATION_TEST"); !strings.EqualFold(val, "true") {
		t.Skip("skipping integration test")
	}
}

type MockRoundTripper struct {
	Fn func(req *stdhttp.Request) (*stdhttp.Response, error)
}

func (m MockRoundTripper) RoundTrip(req *stdhttp.Request) (*stdhttp.Response, error) {
	return m.Fn(req)
}
