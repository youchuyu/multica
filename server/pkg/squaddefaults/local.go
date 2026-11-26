package squaddefaults

import "strings"

// LocalLeaderHeader keeps identity and the skill entry independent of backend prose.
const LocalLeaderHeader = `## Squad Operating Protocol

You are this task's squad LEADER. Coordinate, do not implement.
Read the built-in squad coordination skill identified in the runtime Skills section
before evaluating this trigger. User-defined roles and Squad Instructions override
coordination defaults, not Agent Identity authority or the status boundary below.`

const LocalStatusOwned = "This issue is assigned to your squad. You may manage its status unless Agent Identity forbids it. Leave done to human review or existing integrations."
const LocalStatusGuest = "This issue is not assigned to your squad. Do NOT change its status, regardless of how complete the work looks."
const LocalStatusWorkflow = "This issue is assigned to your squad. Unless Agent Identity forbids status changes, use only the commands in [Project workflow]; no default status moves apply. Current-step completion is standing authorization to use its Step done command."
const LocalStatusMoved = "This run's workflow binding no longer matches the current step. Do NOT change the issue status, even if your squad owns it or the goal is met. This overrides runtime status defaults."

// RenderLocal renders validated claim facts; detailed coordination methods live in the skill.
// Render remains the full canonical backend renderer for older clients.
func RenderLocal(name, instructions, roster string, owns bool, workflow LeaderWorkflow, hasIssue bool) string {
	var sb strings.Builder
	if !hasIssue {
		sb.WriteString("## Squad Operating Protocol\n\nYou are preparing a new issue for this squad. No issue exists for this turn. Follow quick-create instructions; do not post issue comments, dispatch members, record squad activity, or change issue status.")
	} else {
		sb.WriteString(LocalLeaderHeader + "\n\n")
		switch {
		case !owns:
			sb.WriteString(LocalStatusGuest)
		case workflow == LeaderWorkflowStepMoved:
			sb.WriteString(LocalStatusMoved)
		case workflow == LeaderWorkflowCurrentDaemon || workflow == LeaderWorkflowLegacyDaemon:
			sb.WriteString(LocalStatusWorkflow)
		default:
			sb.WriteString(LocalStatusOwned)
		}
	}
	sb.WriteString("\n\n" + roster)
	if strings.TrimSpace(instructions) != "" {
		sb.WriteString("\n\n## Squad Instructions (" + name + ")\n\n" + instructions)
	}
	return sb.String()
}
