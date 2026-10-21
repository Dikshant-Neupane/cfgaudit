// Copyright the cfgaudit contributors
// SPDX-License-Identifier: Apache-2.0

package rules

import (
	"testing"

	"github.com/cfgaudit/cfgaudit/internal/version"
)

// #746: from Claude Code 2.1.299 a single server name given as a plain string
// loads, so it is neither a type mismatch (CFG012) nor a reason to treat the deny
// list as not in force (CFG006). "*" as a string is still dropped upstream.
func TestSingleMcpjsonServerName(t *testing.T) {
	before := &version.Version{Major: 2, Minor: 1, Patch: 298}
	from := &version.Version{Major: 2, Minor: 1, Patch: 299}
	const deny = `,"permissions":{"allow":["Bash(make *)"],"deny":["Bash(rm *)"]}}`

	for _, tc := range []struct {
		name      string
		json      string
		ver       *version.Version
		discarded bool
	}{
		{"enabled name, unknown version", `{"enabledMcpjsonServers":"github"` + deny, nil, false},
		{"disabled name, 2.1.299", `{"disabledMcpjsonServers":"my-server_1"` + deny, from, false},
		{"enabled name, 2.1.298", `{"enabledMcpjsonServers":"github"` + deny, before, true},
		{"wildcard string", `{"enabledMcpjsonServers":"*"` + deny, nil, true},
		{"name with a space", `{"enabledMcpjsonServers":"two words"` + deny, from, true},
		{"array still fine", `{"enabledMcpjsonServers":["github"]` + deny, before, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := settingsTarget(t, tc.json)
			target.ClaudeVersion = tc.ver
			if got := SettingsDiscarded(target.Settings, tc.ver); got != tc.discarded {
				t.Fatalf("SettingsDiscarded = %v, want %v", got, tc.discarded)
			}
			if got := len(CFG012.Check(target)) > 0; got != tc.discarded {
				t.Errorf("CFG012 reported = %v, want %v", got, tc.discarded)
			}
			if got := len(CFG006.Check(target)) > 0; got != tc.discarded {
				t.Errorf("CFG006 reported = %v, want %v (the deny list is in force unless the file is discarded)", got, tc.discarded)
			}
		})
	}
}

func TestDecodeStringList_SingleServerName(t *testing.T) {
	if got := decodeStringList([]byte(`"github"`)); len(got) != 1 || got[0] != "github" {
		t.Errorf(`"github" should read as a one-item list, got %v`, got)
	}
	if got := decodeStringList([]byte(`"*"`)); got != nil {
		t.Errorf(`"*" as a string is not accepted upstream, got %v`, got)
	}
	if got := decodeStringList([]byte(`["a","b"]`)); len(got) != 2 {
		t.Errorf("an array is read as before, got %v", got)
	}
}
