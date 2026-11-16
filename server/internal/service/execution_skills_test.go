package service

import (
	"strings"
	"testing"
)

func TestExecutionSkillsShipWithStableBuiltinIdentity(t *testing.T) {
	skills := (&TaskService{}).BuiltinSkills("", false)
	_, refs := BuildAgentSkillBundles(skills)
	for _, name := range []string{IssueExecutionSkillName, SquadCoordinationSkillName} {
		found := false
		for _, ref := range refs {
			if ref.ID == BuiltinSkillID(name) && ref.Source == "builtin" {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing builtin %s", name)
		}
	}
}

func TestExecutionSkillsKeepSeparateResponsibilities(t *testing.T) {
	for _, tc := range []struct {
		name string
		want []string
	}{
		{IssueExecutionSkillName, []string{"Issue tasks only", "references/context.md", "references/execution.md", "never widens your role", "multica-platform"}},
		{SquadCoordinationSkillName, []string{"squad leader", "no_action", "only if you have not already commented", "Never do both", "handoff registration", "grants no additional"}},
	} {
		skill, ok := findSkill(t, tc.name)
		if !ok {
			t.Fatal(tc.name)
		}
		for _, want := range tc.want {
			if !strings.Contains(skill.Content, want) {
				t.Errorf("%s missing %q", tc.name, want)
			}
		}
		if strings.Contains(skill.Content, "etf-team") {
			t.Error("team-specific contract leaked")
		}
	}
}
