package daemon

import (
	"github.com/multica-ai/multica/server/internal/daemon/execenv"
	"github.com/multica-ai/multica/server/pkg/squaddefaults"
	"os"
	"strings"
	"testing"
)

func TestLegacyCombinedProtocolRendersLocalDefaults(t *testing.T) {
	raw, err := os.ReadFile("testdata/legacy_combined_mul277.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, provider := range []string{"claude", "codex"} {
		task := Task{AgentID: "leader", IssueID: "issue", IsLeaderTask: true, LeaderRoleResolved: true, ProjectID: "project", Agent: &AgentData{ID: "leader", Instructions: string(raw)}}
		prepareLocalExecutionText(&task)
		if task.Agent.Instructions == string(raw) {
			t.Fatal("real legacy protocol not replaced")
		}
		if !strings.Contains(task.Agent.Instructions, squaddefaults.SquadProtocolContext) || strings.Contains(task.Agent.Instructions, "Be terse") || strings.Contains(task.Agent.Instructions, "do not open files") {
			t.Fatal("stale defaults survived")
		}
		for _, v := range []string{"# User Role\nUser text with ## Squad Operating Protocol heading.", "## Squad Roster\n\nLeader (you):", "Exact user team definition.", squaddefaults.SquadParentStatusOwned} {
			if !strings.Contains(task.Agent.Instructions, v) {
				t.Fatalf("changed declaration or permission: %s", v)
			}
		}
		dir := t.TempDir()
		out, err := execenv.InjectRuntimeConfig(dir, provider, execenv.TaskContextForEnv{IssueID: "issue", AgentName: "Leader", AgentID: "leader", AgentInstructions: task.Agent.Instructions})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, task.Agent.Instructions) {
			t.Fatal("local result not delivered")
		}
		again := task.Agent.Instructions
		prepareLocalExecutionText(&task)
		if task.Agent.Instructions != again {
			t.Fatal("not idempotent")
		}
	}
	task := Task{AgentID: "leader", Agent: &AgentData{ID: "leader", Instructions: string(raw)}, LeaderRoleResolved: true}
	prepareLocalExecutionText(&task)
	if task.Agent.Instructions != string(raw) {
		t.Fatal("nonleader changed")
	}
}
