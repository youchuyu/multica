package daemon

import (
	"fmt"
	"github.com/multica-ai/multica/server/pkg/protocol"
	"github.com/multica-ai/multica/server/pkg/squaddefaults"
	"strings"
)

// prepareLocalExecutionText updates presentation only. Missing or unsupported
// optional metadata must not change whether an otherwise runnable task starts.
// Legacy platform templates are replaced only on complete known byte matches;
// user declarations remain opaque. No REST reads or new execution requirements.
func prepareLocalExecutionText(task *Task) {
	if task == nil || task.Agent == nil {
		return
	}
	if taskIsSquadLeader(*task) {
		task.Agent.Instructions, _ = squaddefaults.ReplaceLegacyProtocol(task.Agent.Instructions, task.AgentID)
		task.Agent.SquadBriefing, _ = squaddefaults.ReplaceLegacyProtocol(task.Agent.SquadBriefing, task.AgentID)
	}
	if task.Agent.SquadContext == nil {
		return
	}
	// The renderer commits its replacement only after checking the optional facts.
	// On failure the existing server briefing is preserved verbatim.
	_ = prepareLocalSquadContext(task)
}

// prepareLocalSquadContext never interprets user instructions as transport metadata.
// It replaces any stale separate defaults only after validating server-owned facts.
func prepareLocalSquadContext(task *Task) error {
	if task == nil {
		return fmt.Errorf("local squad context: nil task")
	}
	var facts *protocol.SquadContext
	if task.Agent != nil {
		facts = task.Agent.SquadContext
	}
	leader := task.IsLeaderTask || task.SquadID != ""
	fail := func(reason string) error {
		return fmt.Errorf("local squad context compatibility: %s; optional squad_context v1 cannot be rendered locally; preserve the supplied briefing", reason)
	}
	if !task.LeaderRoleResolved {
		return fail("claim is missing authoritative leader_role_resolved")
	}
	if facts == nil {
		if leader || (task.Agent != nil && task.Agent.SquadBriefing != "") {
			return fail("leader claim is missing squad_context")
		}
		return nil
	}
	if !task.LeaderRoleResolved || !(task.IsLeaderTask || (task.IssueID == "" && task.SquadID != "")) {
		return fail("squad_context does not match the resolved leader role")
	}
	if task.Agent == nil || task.Agent.ID == "" || task.Agent.ID != task.AgentID || facts.LeaderID != task.Agent.ID {
		return fail("squad_context leader_id does not match the claimed agent")
	}
	if facts.SchemaVersion != protocol.SquadContextSchemaV1 {
		return fail("unsupported squad_context schema_version")
	}
	if facts.SquadID == "" || facts.SquadID != task.SquadID || strings.TrimSpace(facts.SquadName) == "" || (task.SquadName != "" && facts.SquadName != task.SquadName) {
		return fail("squad_context squad identity does not match the claim")
	}
	if facts.OwnsParentStatus == nil || strings.TrimSpace(facts.RosterFacts) == "" {
		return fail("squad_context is missing permission or roster facts")
	}
	workflow := squaddefaults.LeaderNoWorkflow
	switch facts.Workflow {
	case protocol.SquadWorkflowNone:
		if task.ProjectWorkflow != nil {
			return fail("squad_context workflow is missing for the project workflow")
		}
	case protocol.SquadWorkflowProject:
		workflow = squaddefaults.LeaderWorkflowCurrentDaemon
		if task.ProjectWorkflow == nil {
			return fail("squad_context project workflow is missing its workflow brief")
		}
	case protocol.SquadWorkflowStepMoved:
		workflow = squaddefaults.LeaderWorkflowStepMoved
		if task.ProjectWorkflow == nil {
			return fail("squad_context moved step is missing its workflow brief")
		}
	default:
		return fail("unsupported squad_context workflow")
	}
	hasIssue := task.IssueID != ""
	if !hasIssue && (*facts.OwnsParentStatus || facts.Workflow != protocol.SquadWorkflowNone) {
		return fail("pre-issue claim grants issue status or workflow authority")
	}
	if hasIssue && *facts.OwnsParentStatus && (task.IssueAssigneeType != "squad" || task.IssueAssigneeID != facts.SquadID) {
		return fail("squad_context status ownership does not match the issue assignee")
	}
	// Local identity and permissions never depend on server-delivered protocol prose.
	task.Agent.SquadBriefing = squaddefaults.RenderLocal(facts.SquadName, facts.UserInstructions, facts.RosterFacts, *facts.OwnsParentStatus, workflow, hasIssue)
	return nil
}
