package service

import "testing"

func TestExecutionPayloadScope(t *testing.T) {
	for _, issue := range []bool{false, true} {
		for _, leader := range []bool{false, true} {
			skills := []AgentSkillData{{Name: IssueExecutionSkillName}, {Name: SquadCoordinationSkillName}, {Name: PlatformSkillName}, {ID: "workspace", Name: SquadCoordinationSkillName}, {Source: "workspace", Name: IssueExecutionSkillName}}
			refs := []AgentSkillRefData{{Source: "builtin", ID: BuiltinSkillID(IssueExecutionSkillName), Name: "allocated-issue"}, {Source: "builtin", ID: BuiltinSkillID(SquadCoordinationSkillName), Name: "allocated-lead"}, {Source: "workspace", ID: BuiltinSkillID(SquadCoordinationSkillName), Name: SquadCoordinationSkillName}}
			got, gotRefs, removed, removedRefs := ScopeExecutionSkills(skills, refs, issue, leader)
			expectedRemoved := 0
			if !issue {
				expectedRemoved++
			}
			if !issue || !leader {
				expectedRemoved++
			}
			if len(got) != 5-expectedRemoved || len(gotRefs) != 3-expectedRemoved || removed != expectedRemoved || removedRefs != expectedRemoved {
				t.Fatalf("issue=%v leader=%v unexpected filtering", issue, leader)
			}
			if got[len(got)-1].Source != "workspace" || gotRefs[len(gotRefs)-1].Source != "workspace" {
				t.Fatal("workspace same-name skill lost")
			}
		}
	}
}

func TestExecutionSkillScope(t *testing.T) {
	for _, issue := range []bool{false, true} {
		for _, leader := range []bool{false, true} {
			for name, want := range map[string]bool{IssueExecutionSkillName: issue, SquadCoordinationSkillName: issue && leader, PlatformSkillName: true, "professional-skill": true} {
				if got := ExecutionSkillApplies(name, issue, leader); got != want {
					t.Errorf("%s issue=%v leader=%v got=%v", name, issue, leader, got)
				}
			}
		}
	}
}
