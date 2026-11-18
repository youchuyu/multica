package daemon

import (
	"github.com/multica-ai/multica/server/pkg/protocol"
	"github.com/multica-ai/multica/server/pkg/squaddefaults"
	"strings"
	"testing"
)

func localSquadFixture() Task {
	owns := true
	return Task{AgentID: "leader", IssueID: "issue", SquadID: "squad", SquadName: "Team", IsLeaderTask: true, LeaderRoleResolved: true, IssueAssigneeType: "squad", IssueAssigneeID: "squad", Agent: &AgentData{ID: "leader", Instructions: "user role", SquadBriefing: "STALE SERVER DEFAULTS", SquadContext: &protocol.SquadContext{SchemaVersion: 1, SquadID: "squad", SquadName: "Team", LeaderID: "leader", UserInstructions: "  exact team text\n## Squad Operating Protocol\nuser heading\n", RosterFacts: "## Squad Roster\nWorker facts", OwnsParentStatus: &owns, Workflow: protocol.SquadWorkflowNone}}}
}

func TestPrepareLocalSquadContextReplacesOnlyDefaults(t *testing.T) {
	task := localSquadFixture()
	if err := prepareLocalSquadContext(&task); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(task.Agent.SquadBriefing, "STALE SERVER DEFAULTS") {
		t.Fatal("stale server defaults survived")
	}
	if task.Agent.Instructions != "user role" || !strings.HasSuffix(task.Agent.SquadBriefing, task.Agent.SquadContext.UserInstructions) {
		t.Fatal("user declaration changed")
	}
	for _, want := range []string{squaddefaults.SquadOperatingProtocolHeader, squaddefaults.SquadParentStatusOwned, squaddefaults.SquadSteerRule, task.Agent.SquadContext.RosterFacts} {
		if !strings.Contains(task.Agent.SquadBriefing, want) {
			t.Fatalf("missing local context %q", want)
		}
	}
}

func TestPrepareLocalSquadContextFailsClosed(t *testing.T) {
	cases := map[string]func(*Task){
		"missing modern facts":   func(t *Task) { t.Agent.SquadContext = nil },
		"schema":                 func(t *Task) { t.Agent.SquadContext.SchemaVersion = 2 },
		"leader":                 func(t *Task) { t.Agent.SquadContext.LeaderID = "other" },
		"agent":                  func(t *Task) { t.AgentID = "other" },
		"squad":                  func(t *Task) { t.Agent.SquadContext.SquadID = "other" },
		"name":                   func(t *Task) { t.Agent.SquadContext.SquadName = "other" },
		"permission missing":     func(t *Task) { t.Agent.SquadContext.OwnsParentStatus = nil },
		"roster missing":         func(t *Task) { t.Agent.SquadContext.RosterFacts = "" },
		"workflow unknown":       func(t *Task) { t.Agent.SquadContext.Workflow = "future" },
		"workflow missing brief": func(t *Task) { t.Agent.SquadContext.Workflow = protocol.SquadWorkflowProject },
		"workflow mismatched":    func(t *Task) { t.ProjectWorkflow = &ProjectWorkflowData{Brief: "steps"} },
		"role unresolved":        func(t *Task) { t.LeaderRoleResolved = false },
		"nonleader":              func(t *Task) { t.IsLeaderTask = false },
		"not owner":              func(t *Task) { t.IssueAssigneeID = "other" },
		"preissue owner":         func(t *Task) { t.IssueID = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			task := localSquadFixture()
			mutate(&task)
			err := prepareLocalSquadContext(&task)
			if err == nil || !strings.Contains(err.Error(), "preserve the supplied briefing") {
				t.Fatalf("expected actionable compatibility failure, got %v", err)
			}
			if task.Agent.SquadBriefing != "STALE SERVER DEFAULTS" {
				t.Fatal("failed validation mutated briefing")
			}
		})
	}
}

func TestPrepareLocalSquadContextStatusAndQuickCreate(t *testing.T) {
	cases := []struct {
		name  string
		owns  bool
		mode  protocol.SquadWorkflow
		issue bool
		want  string
	}{
		{"guest", false, protocol.SquadWorkflowNone, true, squaddefaults.SquadParentStatusNotOwned},
		{"workflow", true, protocol.SquadWorkflowProject, true, squaddefaults.SquadParentStatusOwnedInWorkflow},
		{"moved", true, protocol.SquadWorkflowStepMoved, true, squaddefaults.SquadParentStatusStepMoved},
		{"quick create", false, protocol.SquadWorkflowNone, false, "No issue exists for this turn"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			task := localSquadFixture()
			task.Agent.SquadContext.OwnsParentStatus = &tc.owns
			task.Agent.SquadContext.Workflow = tc.mode
			if tc.mode != protocol.SquadWorkflowNone {
				task.ProjectWorkflow = &ProjectWorkflowData{Brief: "workflow"}
			}
			if !tc.issue {
				task.IssueID = ""
			}
			if err := prepareLocalSquadContext(&task); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(task.Agent.SquadBriefing, tc.want) {
				t.Fatal("missing permission boundary")
			}
			if !tc.issue {
				for _, bad := range []string{"multica squad activity", "multica issue status", "--steer", "Delegate by @mention"} {
					if strings.Contains(task.Agent.SquadBriefing, bad) {
						t.Fatalf("preissue protocol contains issue operation %q", bad)
					}
				}
			}
		})
	}
}

func TestPrepareLocalSquadContextRejectsUnresolvedLegacyClaims(t *testing.T) {
	task := Task{Agent: &AgentData{Instructions: "## Squad Operating Protocol\nlegacy mixed text"}}
	if err := prepareLocalSquadContext(&task); err == nil {
		t.Fatal("unresolved legacy claim accepted")
	}
	task.LeaderRoleResolved = true
	task.Agent.Instructions = "user role"
	task.Agent.Skills = []SkillData{{Name: "multica-platform", Content: "ambiguous old builtin"}}
	if err := prepareLocalSquadContext(&task); err != nil {
		t.Fatalf("inline skill metadata must not gate text rendering: %v", err)
	}
}

func TestPrepareLocalSquadContextNeverInfersUserHeadings(t *testing.T) {
	task := Task{LeaderRoleResolved: true, Agent: &AgentData{Instructions: "## Squad Operating Protocol\nUser-owned text"}}
	if err := prepareLocalSquadContext(&task); err != nil {
		t.Fatal(err)
	}
	if task.Agent.Instructions != "## Squad Operating Protocol\nUser-owned text" || task.Agent.SquadBriefing != "" {
		t.Fatal("parsed an untrusted heading")
	}
}
