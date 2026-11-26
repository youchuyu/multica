package execenv

import "testing"

func TestRuntimeStaticBudget(t *testing.T) {
	for _, provider := range []string{"claude", "codex"} {
		for _, scenario := range []string{"issue", "comment", "leader", "workflow", "project", "chat", "quick", "autopilot"} {
			ctx := TaskContextForEnv{AgentSkills: []SkillContextForEnv{platformSkillFixture(), issueExecutionFixture("multica-issue-execution"), {Name: "multica-squad-coordination", BuiltinName: "multica-squad-coordination"}}}
			ctx.AgentSkills[0].Content += "\n" + platformOperationContractMarker
			switch scenario {
			case "chat":
				ctx.ChatSessionID = "chat"
			case "quick":
				ctx.QuickCreatePrompt = "create"
			case "autopilot":
				ctx.AutopilotRunID = "run"
			default:
				ctx.IssueID = "issue"
			}
			if scenario == "comment" {
				ctx.TriggerCommentID = "comment"
			}
			if scenario == "leader" {
				ctx.IsSquadLeader = true
			}
			if scenario == "workflow" {
				ctx.HasProjectWorkflow = true
			}
			if scenario == "project" {
				ctx.ProjectID = "project"
				ctx.ProjectTitle = "Project"
			}
			if size := len(buildMetaSkillContent(provider, ctx)); size > 3072 {
				t.Errorf("%s/%s platform static rules %d exceed 3072 bytes", provider, scenario, size)
			}
			t.Logf("%s/%s static bytes=%d", provider, scenario, len(buildMetaSkillContent(provider, ctx)))
		}
	}
}
