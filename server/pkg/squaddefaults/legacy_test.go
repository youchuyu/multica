package squaddefaults

import (
	"strings"
	"testing"
)

func TestReplaceLegacyProtocol(t *testing.T) {
	h, _ := legacyTemplates.ReadFile("legacy/combined-v1-header.txt")
	r, _ := legacyTemplates.ReadFile("legacy/combined-v1-hard-rules.txt")
	for _, status := range []string{SquadParentStatusOwned, SquadParentStatusNotOwned, SquadParentStatusOwnedInWorkflow, SquadParentStatusStepMoved} {
		for _, steer := range []string{"", "\n" + SquadSteerRule} {
			prefix := "USER ROLE\n## Squad Operating Protocol\nuser heading\n\n"
			suffix := "\n\n## Squad Roster\n\nLeader (you):\n- Leader — agent — " + "`[@Leader](mention://agent/leader)`" + "\n\nMembers: human\n## Squad Instructions (team)\nUser instructions untouched."
			old := string(h) + "\n" + status + "\n\n" + string(r) + steer
			out, ok := ReplaceLegacyProtocol(prefix+old+suffix, "leader")
			want := prefix + SquadOperatingProtocolHeader + "\n" + status + "\n\n" + SquadOperatingProtocolHardRules + steer + suffix
			if !ok || out != want {
				t.Fatal("replacement changed declarations or failed")
			}
			if next, ok := ReplaceLegacyProtocol(out, "leader"); ok || next != out {
				t.Fatal("not idempotent")
			}
			for _, bad := range []string{prefix + strings.Replace(old, "Be terse", "Be different", 1) + suffix, prefix + old, prefix + old + suffix + "\n\n" + old + suffix, "USER QUOTE " + old + suffix} {
				if next, ok := ReplaceLegacyProtocol(bad, "leader"); ok || next != bad {
					t.Fatal("ambiguous text changed")
				}
			}
			if next, ok := ReplaceLegacyProtocol(prefix+old+suffix, "other"); ok || next != prefix+old+suffix {
				t.Fatal("wrong leader changed")
			}
		}
	}
}
