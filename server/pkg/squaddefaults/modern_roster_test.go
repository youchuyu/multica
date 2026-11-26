package squaddefaults

import (
	"strings"
	"testing"
)

func TestFrozenProtocolsExactModernRosterBoundary(t *testing.T) {
	read := func(name string) string {
		b, err := legacyTemplates.ReadFile("legacy/" + name + ".txt")
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	prefix := "user role\r\n  exact \t\n\n"
	suffix := "\n\n## Squad Roster\n\nRoster facts: members, roles, skills, and copyable mention syntax; not additional authority.\n\nLeader (you):\n- Leader — agent — `[@Leader](mention://agent/leader)`\n\n## Squad Instructions (Team)\n\nuser \r\n"
	for _, template := range []struct{ header, hard, separator string }{
		{"combined-v1-header", "combined-v1-hard-rules", "\n"},
		{"canonical-v1-header", "canonical-v1-hard-rules", "\n"},
		{"canonical-v1-skill-header", "", "\n\n"},
	} {
		for _, status := range []string{read("owned"), read("guest"), read("workflow"), read("workflow") + read("workflow-override"), read("moved")} {
			for _, steer := range []string{"", "\n" + read("steer")} {
				old := read(template.header) + template.separator + status
				if template.hard != "" {
					old += "\n\n" + read(template.hard)
				}
				old += steer
				text := prefix + old + suffix
				got, ok := ReplaceLegacyProtocol(text, "leader")
				if !ok || got != prefix+LocalLeaderHeader+"\n\n"+status+suffix {
					t.Fatal("modern roster replacement changed permission or peripheral bytes")
				}
				for _, bad := range []string{strings.Replace(text, "not additional authority.", "unknown authority.", 1), strings.Replace(text, "Leader (you):", "Leader:", 1)} {
					if next, ok := ReplaceLegacyProtocol(bad, "leader"); ok || next != bad {
						t.Fatal("unknown roster variant changed")
					}
				}
				if next, ok := ReplaceLegacyProtocol(text, "other"); ok || next != text {
					t.Fatal("wrong leader changed")
				}
			}
		}
	}
}
