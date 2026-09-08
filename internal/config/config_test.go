// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package config

import "testing"

const (
	testPath = "../../testdata"
	testFile = "config.toml"
)

func TestNew(t *testing.T) {
	conf, err := New(testPath, testFile)
	if err != nil {
		t.Fatalf("failed to create config: %s", err)
	}
	if conf == nil {
		t.Fatalf("config is nil")
	}
}
