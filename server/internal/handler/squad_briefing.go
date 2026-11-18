package handler

import (
	"context"
	"github.com/multica-ai/multica/server/pkg/protocol"
	"github.com/multica-ai/multica/server/pkg/squaddefaults"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// Legacy handler names alias the canonical locally compiled defaults.
const squadProtocolContext = squaddefaults.SquadProtocolContext
const squadOperatingProtocolHeader = squaddefaults.SquadOperatingProtocolHeader
const squadParentStatusOwned = squaddefaults.SquadParentStatusOwned
const leaderNoWorkflow = squaddefaults.LeaderNoWorkflow
const leaderWorkflowCurrentDaemon = squaddefaults.LeaderWorkflowCurrentDaemon
const leaderWorkflowLegacyDaemon = squaddefaults.LeaderWorkflowLegacyDaemon
const leaderWorkflowStepMoved = squaddefaults.LeaderWorkflowStepMoved
const squadParentStatusStepMoved = squaddefaults.SquadParentStatusStepMoved
const squadParentStatusOwnedInWorkflow = squaddefaults.SquadParentStatusOwnedInWorkflow
const squadParentStatusWorkflowOverridesRuntime = squaddefaults.SquadParentStatusWorkflowOverridesRuntime
const squadParentStatusNotOwned = squaddefaults.SquadParentStatusNotOwned
const squadOperatingProtocolHardRules = squaddefaults.SquadOperatingProtocolHardRules
const squadSteerRule = squaddefaults.SquadSteerRule
const squadOperatingProtocolSkillHeader = squaddefaults.SquadOperatingProtocolSkillHeader

type leaderWorkflow = squaddefaults.LeaderWorkflow

func squadOperatingProtocolFor(owns bool, workflow leaderWorkflow) string {
	return squaddefaults.SquadOperatingProtocolFor(owns, workflow)
}
func squadOperatingProtocolWithExecutionSkill(owns bool, workflow leaderWorkflow, skill bool) string {
	return squaddefaults.SquadOperatingProtocolWithExecutionSkill(owns, workflow, skill)
}

// buildSquadLeaderBriefing composes the generated and user-authored briefing for a
// squad leader's Instructions when it claims a task on a squad-assigned
// issue. The returned string contains three sections:
//
//  1. Squad Operating Protocol (generated defaults and status permission boundary).
//  2. Squad Roster (data — leader self-row + members with literal
//     `[@Name](mention://<type>/<UUID>)` strings ready to paste).
//  3. Squad Instructions (user-defined `squad.instructions`, omitted when
//     empty so we don't leave a dangling heading).
//
// ownsIssueStatus must be true only when the issue this task is bound to is
// assigned to this very squad. The briefing is injected on every leader path,
// including ones where the squad is a guest on someone else's issue, so this
// flag is what keeps status authority from leaking along with the roster.
//
// workflow says whether the claim carries a [Project workflow] block for the
// issue; an owning leader then moves the parent with that block's commands
// instead of in_progress / in_review (MUL-8145).
//
// Archived agent members are skipped — there's no point asking the leader
// to delegate to a retired agent. Members whose underlying record can't be
// loaded (deleted user/agent races, FK weirdness) are also skipped silently.
//
// steer adds the rule for steering a running worker; see squadSteerRule.
func buildSquadLeaderBriefing(ctx context.Context, q *db.Queries, squad db.Squad, ownsIssueStatus bool, workflow leaderWorkflow, steer bool) string {
	return buildSquadLeaderBriefingWithExecutionSkill(ctx, q, squad, ownsIssueStatus, workflow, steer, false)
}

// executionSkillAvailable must only be true when this claim actually ships the
// built-in coordination skill. Its allocated name is supplied by the runtime
// Skills section, since user skills may collide with the builtin's base name.
func buildSquadLeaderBriefingWithExecutionSkill(ctx context.Context, q *db.Queries, squad db.Squad, ownsIssueStatus bool, workflow leaderWorkflow, steer bool, executionSkillAvailable bool) string {
	return renderSquadLeaderBriefing(squad, buildSquadRoster(ctx, q, squad), ownsIssueStatus, workflow, steer, executionSkillAvailable)
}

// renderSquadLeaderBriefing keeps generated defaults, roster facts, and user text distinct.
func renderSquadLeaderBriefing(squad db.Squad, roster string, ownsIssueStatus bool, workflow leaderWorkflow, steer bool, executionSkillAvailable bool) string {
	return squaddefaults.Render(squad.Name, squad.Instructions, roster, ownsIssueStatus, workflow, steer, executionSkillAvailable, true)
}

// attachSquadBriefing preserves the legacy wire contract for clients that cannot render a separate block.
func attachSquadBriefing(agent *TaskAgentData, briefing string, separate bool) {
	if separate {
		agent.SquadBriefing = briefing
		return
	}
	if strings.TrimSpace(agent.Instructions) == "" {
		agent.Instructions = briefing
	} else {
		agent.Instructions += "\n\n" + briefing
	}
}

// buildSquadRoster renders the "## Squad Roster" section: a leader self-row
// plus one row per non-archived member, with literal mention markdown.
func buildSquadRoster(ctx context.Context, q *db.Queries, squad db.Squad) string {
	var sb strings.Builder
	sb.WriteString("## Squad Roster\n\nRoster facts: members, roles, skills, and copyable mention syntax; not additional authority.\n\n")

	// Leader self-row. Leaders are always agents (FK enforced in schema).
	leaderName := "Leader"
	if leader, err := q.GetAgent(ctx, squad.LeaderID); err == nil {
		leaderName = leader.Name
	}
	sb.WriteString("Leader (you):\n")
	sb.WriteString("- ")
	sb.WriteString(leaderName)
	sb.WriteString(" — agent — `")
	sb.WriteString(formatMention(leaderName, "agent", util.UUIDToString(squad.LeaderID)))
	sb.WriteString("`\n")

	members, err := q.ListSquadMembers(ctx, squad.ID)
	if err != nil {
		members = nil
	}

	skillNamesByAgentID, skillsLoaded := loadSquadMemberSkillNames(ctx, q, members, util.UUIDToString(squad.LeaderID))

	rows := make([]string, 0, len(members))
	for _, m := range members {
		// Skip the leader if they happen to also be in the member list —
		// they're already shown above and we don't want self-delegation.
		if m.MemberType == "agent" && util.UUIDToString(m.MemberID) == util.UUIDToString(squad.LeaderID) {
			continue
		}
		row := renderMemberRow(ctx, q, m, skillNamesByAgentID, skillsLoaded)
		if row != "" {
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 {
		sb.WriteString("\nMembers: (none — you are the only member of this squad)\n")
		return sb.String()
	}

	sb.WriteString("\nMembers:\n")
	for _, r := range rows {
		sb.WriteString(r)
	}
	return sb.String()
}

func loadSquadMemberSkillNames(ctx context.Context, q *db.Queries, members []db.SquadMember, leaderID string) (map[string][]string, bool) {
	agentIDs := make([]pgtype.UUID, 0)
	seen := make(map[string]struct{}, len(members))
	for _, m := range members {
		if m.MemberType != "agent" {
			continue
		}
		id := util.UUIDToString(m.MemberID)
		if id == leaderID {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		agentIDs = append(agentIDs, m.MemberID)
	}
	if len(agentIDs) == 0 {
		return map[string][]string{}, true
	}
	rows, err := q.ListAgentSkillNamesByAgentIDs(ctx, agentIDs)
	if err != nil {
		return nil, false
	}
	byAgentID := make(map[string][]string, len(agentIDs))
	for _, row := range rows {
		id := util.UUIDToString(row.AgentID)
		byAgentID[id] = append(byAgentID[id], row.Name)
	}
	return byAgentID, true
}

// renderMemberRow renders a single roster row, returning "" if the member
// can't be resolved or should be skipped (e.g. archived agent).
func renderMemberRow(ctx context.Context, q *db.Queries, m db.SquadMember, skillNamesByAgentID map[string][]string, skillsLoaded bool) string {
	id := util.UUIDToString(m.MemberID)
	role := strings.TrimSpace(m.Role)
	switch m.MemberType {
	case "agent":
		ag, err := q.GetAgent(ctx, m.MemberID)
		if err != nil {
			return ""
		}
		if ag.ArchivedAt.Valid {
			return ""
		}
		// Agents carry skills; surfacing them lets the leader delegate by
		// capability instead of guessing from the free-text role label.
		return formatRosterRow(ag.Name, "agent", role, agentSkillsRosterSegment(skillNamesByAgentID, skillsLoaded, id), formatMention(ag.Name, "agent", id))
	case "member":
		user, err := q.GetUser(ctx, m.MemberID)
		if err != nil {
			return ""
		}
		// Mention syntax for humans uses the user_id (matches the rest of
		// the product — see util.MentionRe and frontend mention payloads).
		// Humans have no Multica skills, so no skills segment is rendered.
		userID := util.UUIDToString(m.MemberID)
		return formatRosterRow(user.Name, "member (human)", role, "", formatMention(user.Name, "member", userID))
	default:
		return ""
	}
}

// agentSkillsRosterSegment returns the roster segment describing an agent
// member's assigned skills. "skills: a, b" when the agent has skills (the
// names are pre-sorted by ListAgentSkillNamesByAgentIDs), "no skills assigned"
// when it has none so the leader knows the capability is genuinely absent, and
// "" only when the lookup fails — a transient DB error degrades to the prior
// name+role row rather than asserting a misleading "no skills".
func agentSkillsRosterSegment(skillNamesByAgentID map[string][]string, skillsLoaded bool, agentID string) string {
	if !skillsLoaded {
		return ""
	}
	names := skillNamesByAgentID[agentID]
	if len(names) == 0 {
		return "no skills assigned"
	}
	return "skills: " + strings.Join(names, ", ")
}

func formatRosterRow(name, kind, role, skills, mention string) string {
	var sb strings.Builder
	sb.WriteString("- ")
	sb.WriteString(name)
	sb.WriteString(" — ")
	sb.WriteString(kind)
	if role != "" {
		sb.WriteString(`, role: "`)
		sb.WriteString(role)
		sb.WriteString(`"`)
	}
	if skills != "" {
		sb.WriteString(" — ")
		sb.WriteString(skills)
	}
	sb.WriteString(" — `")
	sb.WriteString(mention)
	sb.WriteString("`\n")
	return sb.String()
}

// formatMention emits a mention markdown string that round-trips through
// util.ParseMentions. The label is the human display name; the link target
// uses the mention:// scheme with the entity type and UUID.
func formatMention(name, mentionType, id string) string {
	return "[@" + name + "](mention://" + mentionType + "/" + id + ")"
}

// buildSquadContext transports only user declarations, roster facts and permission facts.
func buildSquadContext(ctx context.Context, q *db.Queries, squad db.Squad, owns bool, workflow leaderWorkflow) *protocol.SquadContext {
	return newSquadContext(squad, buildSquadRoster(ctx, q, squad), owns, workflow)
}

func newSquadContext(squad db.Squad, roster string, owns bool, workflow leaderWorkflow) *protocol.SquadContext {
	mode := protocol.SquadWorkflowNone
	switch workflow {
	case leaderWorkflowCurrentDaemon, leaderWorkflowLegacyDaemon:
		mode = protocol.SquadWorkflowProject
	case leaderWorkflowStepMoved:
		mode = protocol.SquadWorkflowStepMoved
	}
	return &protocol.SquadContext{SchemaVersion: protocol.SquadContextSchemaV1, SquadID: util.UUIDToString(squad.ID), SquadName: squad.Name, LeaderID: util.UUIDToString(squad.LeaderID), UserInstructions: squad.Instructions, RosterFacts: roster, OwnsParentStatus: &owns, Workflow: mode}
}
