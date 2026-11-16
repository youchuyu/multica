package execenv

import (
	"strings"
	"testing"
)

func TestRuntimeDefinitionLayout(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "dsh"} {
		for _, kind := range []string{"issue", "comment", "chat", "autopilot", "quick-create"} {
			for _, leader := range []bool{false, true} {
				t.Run(provider+"/"+kind+"/"+map[bool]string{false: "worker", true: "leader"}[leader], func(t *testing.T) {
					instructions := "  Role instructions\n\n## Squad Operating Protocol\nThis heading is authored content.\n  "
					ctx := TaskContextForEnv{AgentName: "Executor", AgentID: "agent-id", AgentInstructions: instructions, IsSquadLeader: leader}
					switch kind {
					case "issue":
						ctx.IssueID = "issue"
					case "comment":
						ctx.IssueID = "issue"
						ctx.TriggerCommentID = "comment"
					case "chat":
						ctx.ChatSessionID = "chat"
					case "autopilot":
						ctx.AutopilotRunID = "run"
					case "quick-create":
						ctx.QuickCreatePrompt = "create"
					}
					if leader {
						ctx.SquadBriefing = "## Squad Operating Protocol\nOriginal protocol\n\n## Squad Roster\nOriginal roster\n\n## Squad Instructions (team)\n  Original instructions  \n"
					}
					out := buildMetaSkillContent(provider, ctx)
					definition := strings.Index(out, "# Executor Definition\n")
					if definition <= strings.Index(out, "## Output\n") {
						t.Fatal("agent definition must follow all platform sections")
					}
					if strings.Count(out, instructions) != 1 {
						t.Fatal("agent instructions changed or duplicated")
					}
					if !strings.Contains(out, "where they conflict, follow those definitions") || !strings.Contains(out, "Platform runtime mechanisms and permission boundaries still apply") {
						t.Fatal("missing definition baseline or platform boundary")
					}
					if leader {
						if strings.Index(out, "# Team Information\n") <= definition || !strings.HasSuffix(out, ctx.SquadBriefing+"\n\n") || strings.Count(out, ctx.SquadBriefing) != 1 {
							t.Fatal("complete team briefing must remain unchanged at the end")
						}
					} else if strings.Contains(out, "# Team Information\n") {
						t.Fatal("worker must not gain a team briefing from authored headings")
					}
				})
			}
		}
	}
}

func TestRuntimeEmptyAndLegacyDefinitions(t *testing.T) {
	for _, ctx := range []TaskContextForEnv{{}, {AgentName: "Named"}, {AgentInstructions: "Legacy role\n\n## Squad Operating Protocol\nLegacy combined briefing"}, {SquadBriefing: "Opaque squad only"}} {
		out := buildMetaSkillContent("claude", ctx)
		if ctx.AgentInstructions != "" && !strings.HasSuffix(out, ctx.AgentInstructions+"\n\n") {
			t.Fatal("legacy combined instructions must remain opaque")
		}
		if ctx.SquadBriefing == "" && strings.Contains(out, "# Team Information\n") {
			t.Fatal("empty squad gained a section")
		}
		if ctx.AgentName == "" && ctx.AgentInstructions == "" && strings.Contains(out, "# Agent Definition\n") {
			t.Fatal("empty agent gained a definition")
		}
	}
}
