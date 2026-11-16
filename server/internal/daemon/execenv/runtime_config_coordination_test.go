package execenv

import (
	"strings"
	"testing"
)

func TestSquadCoordinationAllocatedEntryIsLeaderOnly(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "traecli"} {
		for _, leader := range []bool{false, true} {
			ctx := TaskContextForEnv{IssueID: "issue", IsSquadLeader: leader, AgentSkills: []SkillContextForEnv{{Name: "allocated-coordination", BuiltinName: "multica-squad-coordination"}}}
			out := buildMetaSkillContent(provider, ctx)
			got := strings.Contains(out, "Before coordinating this squad-leader turn, read the `allocated-coordination` skill")
			if leader && strings.Count(out, "Before coordinating this squad-leader turn") != 1 {
				t.Fatal("coordination entry must occur once")
			}
			if got != leader {
				t.Errorf("provider=%s leader=%v entry=%v", provider, leader, got)
			}
		}
	}
}
