package daemon

import (
	"reflect"
	"testing"

	"github.com/multica-ai/multica/server/pkg/executiondefaults"
)

func TestNormalizeEquivalentLegacyExecutionSkills(t *testing.T) {
	for _, local := range executiondefaults.Skills() {
		old := SkillData{Name: local.Name, Content: local.Content, Description: local.Description}
		for _, f := range local.Files {
			old.Files = append(old.Files, SkillFileData{Path: f.Path, Content: f.Content})
		}
		for _, tc := range []struct {
			name   string
			change func(*SkillData)
			match  bool
		}{
			{"exact", func(*SkillData) {}, true},
			{"user UUID", func(s *SkillData) { s.ID = "user-id" }, false},
			{"plugin", func(s *SkillData) { s.Source = "plugin" }, false},
			{"changed body", func(s *SkillData) { s.Content += "x" }, false},
			{"extra file", func(s *SkillData) { s.Files = append(s.Files, SkillFileData{Path: "extra", Content: "x"}) }, false},
		} {
			t.Run(local.Name+"/"+tc.name, func(t *testing.T) {
				s := old
				s.Files = append([]SkillFileData(nil), old.Files...)
				tc.change(&s)
				task := Task{Agent: &AgentData{Skills: []SkillData{s}}}
				normalizeEquivalentLegacyExecutionSkills(&task)
				if executiondefaults.Owns(task.Agent.Skills[0].Source, task.Agent.Skills[0].ID) != tc.match {
					t.Fatal("incorrect anonymous classification")
				}
			})
		}
	}
}

func TestFilterLocalExecutionSkillRefsUsesIdentity(t *testing.T) {
	id := executiondefaults.BuiltinSkillID(executiondefaults.PlatformSkillName)
	keep := []SkillRefData{
		{Source: "workspace", ID: id, Name: executiondefaults.PlatformSkillName},
		{Source: "plugin", ID: id},
		{Source: "", ID: id},
		{Source: "builtin", ID: "builtin:multica-onboarding"},
		{Source: "builtin", ID: "builtin:future-protocol", Name: executiondefaults.PlatformSkillName},
	}
	task := Task{Agent: &AgentData{SkillRefs: append([]SkillRefData{{Source: "builtin", ID: id, Name: "allocated-collision"}}, keep...)}}
	filterLocalExecutionSkillRefs(&task)
	if !reflect.DeepEqual(task.Agent.SkillRefs, keep) {
		t.Fatalf("refs changed: %+v", task.Agent.SkillRefs)
	}
	filterLocalExecutionSkillRefs(nil)
	filterLocalExecutionSkillRefs(&Task{})
}

func TestMergeLocalExecutionSkillsScopesAndPreservesUsers(t *testing.T) {
	for _, tc := range []struct {
		name, issue string
		leader      bool
		count       int
	}{
		{"chat", "", false, 1}, {"issue", "issue-1", false, 2}, {"leader", "issue-1", true, 3}, {"non-issue-leader", "", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			users := []SkillData{
				{ID: "user", Source: "workspace", Name: executiondefaults.PlatformSkillName, Content: "user content"},
				{ID: executiondefaults.BuiltinSkillID(executiondefaults.IssueExecutionSkillName), Source: "workspace", Name: executiondefaults.IssueExecutionSkillName, Content: "identity collision"},
				{Name: executiondefaults.PlatformSkillName, Content: "ambiguous historical user skill"},
				{ID: "builtin:multica-onboarding", Source: "builtin", Name: "multica-onboarding", Content: "remote system procedure"},
			}
			task := Task{IssueID: tc.issue, LeaderRoleResolved: true, IsLeaderTask: tc.leader, Agent: &AgentData{Skills: append([]SkillData(nil), users...)}}
			for _, local := range executiondefaults.Skills() {
				task.Agent.Skills = append(task.Agent.Skills, SkillData{ID: local.ID, Source: local.Source, Name: "allocated-" + local.Name, Content: "stale remote content"})
			}
			mergeLocalExecutionSkills(&task)
			if len(task.Agent.Skills) != len(users)+tc.count {
				t.Fatalf("got %d skills", len(task.Agent.Skills))
			}
			if !reflect.DeepEqual(task.Agent.Skills[:len(users)], users) {
				t.Fatal("user/system-only content changed")
			}
			for _, local := range task.Agent.Skills[len(users):] {
				if !executiondefaults.Owns(local.Source, local.ID) || local.Content == "stale remote content" || local.Hash == "" || local.SizeBytes == 0 {
					t.Fatalf("invalid local skill: %+v", local)
				}
				for _, file := range local.Files {
					if file.SHA256 == "" || file.SizeBytes == 0 {
						t.Fatalf("supporting metadata missing: %s", file.Path)
					}
				}
			}
			before := append([]SkillData(nil), task.Agent.Skills...)
			mergeLocalExecutionSkills(&task)
			if !reflect.DeepEqual(task.Agent.Skills, before) {
				t.Fatal("merge not idempotent")
			}
		})
	}
	mergeLocalExecutionSkills(nil)
	mergeLocalExecutionSkills(&Task{})
}

func TestLocalExecutionSkillScopeHonorsResolvedMemberRole(t *testing.T) {
	task := Task{IssueID: "issue", LeaderRoleResolved: true, Agent: &AgentData{Instructions: squadBriefingMarker}}
	mergeLocalExecutionSkills(&task)
	for _, skill := range task.Agent.Skills {
		if skill.ID == executiondefaults.BuiltinSkillID(executiondefaults.SquadCoordinationSkillName) {
			t.Fatal("display marker overrode authoritative member role")
		}
	}
}
