package execenv

import (
	"strings"
	"testing"
)

func TestPlatformOperationDisclosureMatrix(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "traecli"} {
		for _, surface := range []string{"issue", "chat", "quick", "autopilot"} {
			for _, state := range []string{"current", "old", "hidden", "impostor", "absent", "renamed"} {
				t.Run(provider+"/"+surface+"/"+state, func(t *testing.T) {
					ctx := TaskContextForEnv{AgentName: "Role", AgentInstructions: "Never create a PR."}
					switch surface {
					case "issue":
						ctx.IssueID = "issue"
					case "chat":
						ctx.ChatSessionID = "chat"
					case "quick":
						ctx.QuickCreatePrompt = "create"
					case "autopilot":
						ctx.AutopilotRunID = "run"
					}
					skill := platformSkillFixture()
					if state != "old" {
						skill.Content += "\n" + platformOperationContractMarker
					}
					if state == "renamed" {
						skill.Name = "allocated-platform"
					}
					if state == "impostor" {
						skill.BuiltinName = ""
					}
					if state == "hidden" {
						skill.Content = "---\nname: multica-platform\ndisable-model-invocation: true\n---\n" + platformOperationContractMarker
					}
					if state != "absent" {
						ctx.AgentSkills = []SkillContextForEnv{skill}
					}
					brief := buildMetaSkillContent(provider, ctx)
					reduced := state == "current" || state == "renamed"
					if strings.Contains(brief, "## Issue Body Formatting") == reduced {
						t.Fatal("formatting fallback gate mismatch")
					}
					requirePhrases(t, brief, "Never create a PR.", "MULTICA_CLI_PATH", "Runtime-local paths are never deliverables")
					if reduced && surface != "issue" {
						requirePhrases(t, brief, "Before acting, load the installed contract(s)", skill.Name+" (platform actions)", "JSON stdout and diagnostic stderr are separate", "Skill availability does not authorize mutations")
						if strings.Contains(brief, "[--due-date") {
							t.Fatal("command catalog duplicated")
						}
						if surface == "issue" {
							requirePhrases(t, brief, "--content-file", "Read the platform issues contract before posting")
							if strings.Contains(brief, "Gate the cleanup on the post succeeding") {
								t.Fatal("receipt method duplicated")
							}
						}
					} else if surface == "issue" && !reduced {
						requirePhrases(t, brief, "Gate the cleanup on the post succeeding", "empty stdout alone does not prove success")
					}
				})
			}
		}
	}
}
