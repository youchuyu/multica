package squaddefaults

import (
	"strings"
	"testing"
)

func TestRenderLocalPermissionsAndOpaqueFacts(t *testing.T) {
	roster := "## Squad Roster\n\nexact roster \t\r\n"
	user := "  ## Squad Operating Protocol\r\nuser text\t\n"
	for _, tc := range []struct {
		name     string
		owns     bool
		workflow LeaderWorkflow
		want     string
	}{
		{"owner", true, LeaderNoWorkflow, LocalStatusOwned},
		{"guest", false, LeaderNoWorkflow, LocalStatusGuest},
		{"workflow", true, LeaderWorkflowCurrentDaemon, LocalStatusWorkflow},
		{"legacy workflow", true, LeaderWorkflowLegacyDaemon, LocalStatusWorkflow},
		{"moved", true, LeaderWorkflowStepMoved, LocalStatusMoved},
		{"workflow guest", false, LeaderWorkflowCurrentDaemon, LocalStatusGuest},
		{"moved guest", false, LeaderWorkflowStepMoved, LocalStatusGuest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderLocal("Team", user, roster, tc.owns, tc.workflow, true)
			want := LocalLeaderHeader + "\n\n" + tc.want + "\n\n" + roster + "\n\n## Squad Instructions (Team)\n\n" + user
			if got != want {
				t.Fatal("changed identity, authority, roster or user bytes")
			}
		})
	}
}

func TestRenderLocalSizeAndCanonicalCompatibility(t *testing.T) {
	canonical := Render("Team", "", "", true, LeaderNoWorkflow, true, false, true)
	local := RenderLocal("Team", "", "", true, LeaderNoWorkflow, true)
	t.Logf("owner generated briefing: canonical=%d bytes, local=%d bytes, reduction=%.1f%%", len(canonical), len(local), 100*(1-float64(len(local))/float64(len(canonical))))
	if len(local)*4 >= len(canonical) {
		t.Fatal("local leader briefing exceeded quarter of canonical size")
	}
	for _, want := range []string{SquadOperatingProtocolHeader, SquadParentStatusOwned, SquadOperatingProtocolHardRules, SquadSteerRule} {
		if !strings.Contains(canonical, want) {
			t.Fatal("backend full canonical renderer changed")
		}
	}
}

func TestReplaceFrozenCanonicalProtocols(t *testing.T) {
	read := func(name string) string {
		b, err := legacyTemplates.ReadFile("legacy/" + name + ".txt")
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	prefix := "role\r\n  opaque \t\n\n"
	suffix := "\n\n## Squad Roster\n\nLeader (you):\n- Leader — agent — `[@Leader](mention://agent/leader)`\n\n## Squad Instructions (T)\n\n  exact user\r\n"
	for _, skill := range []bool{false, true} {
		for _, status := range []string{read("owned"), read("guest"), read("workflow"), read("workflow") + read("workflow-override"), read("moved")} {
			for _, steer := range []string{"", "\n" + read("steer")} {
				old := read("canonical-v1-header") + "\n" + status + "\n\n" + read("canonical-v1-hard-rules")
				if skill {
					old = read("canonical-v1-skill-header") + "\n\n" + status
				}
				old += steer
				text := prefix + old + suffix
				got, ok := ReplaceLegacyProtocol(text, "leader")
				if !ok || got != prefix+LocalLeaderHeader+"\n\n"+status+suffix {
					t.Fatal("canonical replacement changed permission or peripheral bytes")
				}
				if next, ok := ReplaceLegacyProtocol(got, "leader"); ok || next != got {
					t.Fatal("replacement not idempotent")
				}
				for _, bad := range []string{strings.Replace(text, "Coordinate", "Unknown", 1), text + "\n\n" + old + suffix, prefix + old, "quote " + old + suffix} {
					// Full templates use a different first sentence than skill templates.
					if bad == text {
						bad = strings.Replace(text, "Your job", "Unknown job", 1)
					}
					if next, ok := ReplaceLegacyProtocol(bad, "leader"); ok || next != bad {
						t.Fatal("unknown or ambiguous variant changed")
					}
				}
				if next, ok := ReplaceLegacyProtocol(text, "other"); ok || next != text {
					t.Fatal("wrong leader changed")
				}
			}
		}
	}
}
