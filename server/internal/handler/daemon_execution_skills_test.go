package handler

import (
	"github.com/multica-ai/multica/server/internal/service"
	"testing"
)

func TestClaimExecutionSkillUsesBuiltinIdentity(t *testing.T) {
	name := service.SquadCoordinationSkillName
	cases := []struct {
		name  string
		agent *TaskAgentData
		want  bool
	}{
		{"missing", nil, false},
		{"workspace-same-name", &TaskAgentData{Skills: []service.AgentSkillData{{ID: "workspace-id", Name: name}}}, false},
		{"inline-builtin", &TaskAgentData{Skills: []service.AgentSkillData{{Name: name}}}, true},
		{"builtin-ref", &TaskAgentData{SkillRefs: []service.AgentSkillRefData{{Source: "builtin", ID: service.BuiltinSkillID(name), Name: "allocated"}}}, true},
		{"workspace-ref", &TaskAgentData{SkillRefs: []service.AgentSkillRefData{{Source: "workspace", ID: service.BuiltinSkillID(name), Name: name}}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := claimHasBuiltinSkill(c.agent, name); got != c.want {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}
