package execenv

import (
	"strings"
	"testing"
)

func TestDefinitionConflictContracts(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "traecli", "dsh"} {
		for _, surface := range []string{"issue", "chat", "quick-create", "autopilot"} {
			t.Run(provider+"/"+surface, func(t *testing.T) {
				ctx := TaskContextForEnv{AgentName: "Reviewer", AgentInstructions: "Never change issue status. Delegate only; do not implement.", SquadBriefing: "## Squad Instructions (review)\nUse the team review process."}
				switch surface {
				case "issue":
					ctx.IssueID = "issue"
				case "chat":
					ctx.ChatSessionID = "chat"
				case "quick-create":
					ctx.QuickCreatePrompt = "Create an issue"
				case "autopilot":
					ctx.AutopilotRunID = "run"
				}
				out := buildMetaSkillContent(provider, ctx)
				requirePhrases(t, out, ctx.AgentInstructions, ctx.SquadBriefing,
					"when they conflict with Agent Identity, follow Agent Identity",
					"current task supplies this turn's goal within those boundaries",
					"document and built-in skills supplement",
					"task kind and injected status authority are not optional defaults",
					"default status step does not override a prohibition on status writes",
					"request to implement does not override a delegation-only role")
				if strings.Count(out, "## Definition Boundaries\n") != 1 {
					t.Fatal("conflict contract must have one owner")
				}
				if strings.Index(out, "## Definition Boundaries\n") >= strings.Index(out, "# Reviewer Definition\n") {
					t.Fatal("definitions must follow their boundary contract")
				}
			})
		}
	}
}

func TestCoordinationEntryRequiresIssueSurface(t *testing.T) {
	for _, ctx := range []TaskContextForEnv{{QuickCreatePrompt: "create"}, {ChatSessionID: "chat"}, {AutopilotRunID: "run"}} {
		ctx.IsSquadLeader = true
		ctx.AgentSkills = []SkillContextForEnv{{Name: "allocated-coordination", BuiltinName: "multica-squad-coordination"}}
		for _, provider := range []string{"claude", "codex", "traecli"} {
			out := buildMetaSkillContent(provider, ctx)
			if strings.Contains(out, "Before coordinating this squad-leader turn") {
				t.Fatalf("%s routed issue coordination to an issue-less surface", provider)
			}
		}
	}
}
