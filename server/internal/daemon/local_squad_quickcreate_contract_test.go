package daemon

import (
	"github.com/multica-ai/multica/server/pkg/protocol"
	"strings"
	"testing"
)

// Quick-create historically resolves the squad through picker identity rather
// than is_leader_task. Modern context must preserve that explicit wire contract.
func TestPrepareLocalSquadContextQuickCreateResolvedPickerRole(t *testing.T) {
	task := localSquadFixture()
	task.IssueID = ""
	task.IsLeaderTask = false
	owns := false
	task.Agent.SquadContext.OwnsParentStatus = &owns
	if err := prepareLocalSquadContext(&task); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(task.Agent.SquadBriefing, "No issue exists for this turn") {
		t.Fatal("missing pre-issue protocol")
	}
	for _, bad := range []string{"multica squad activity", "multica issue status", "--steer", "Delegate by @mention"} {
		if strings.Contains(task.Agent.SquadBriefing, bad) {
			t.Fatalf("quick-create includes issue operation %q", bad)
		}
	}
}

func TestPrepareLocalSquadContextPickerRoleRequiresPreIssueAndResolvedIdentity(t *testing.T) {
	cases := map[string]func(*Task){
		"issue guest cannot infer leader from picker": func(task *Task) { task.IsLeaderTask = false },
		"unresolved picker": func(task *Task) {
			task.IssueID = ""
			task.IsLeaderTask = false
			task.LeaderRoleResolved = false
			owns := false
			task.Agent.SquadContext.OwnsParentStatus = &owns
		},
		"missing picker identity": func(task *Task) {
			task.IssueID = ""
			task.IsLeaderTask = false
			task.SquadID = ""
			owns := false
			task.Agent.SquadContext.OwnsParentStatus = &owns
		},
		"preissue workflow": func(task *Task) {
			task.IssueID = ""
			task.IsLeaderTask = false
			owns := false
			task.Agent.SquadContext.OwnsParentStatus = &owns
			task.Agent.SquadContext.Workflow = protocol.SquadWorkflowProject
			task.ProjectWorkflow = &ProjectWorkflowData{Brief: "steps"}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			task := localSquadFixture()
			mutate(&task)
			if err := prepareLocalSquadContext(&task); err == nil {
				t.Fatal("invalid picker role accepted")
			}
		})
	}
}
