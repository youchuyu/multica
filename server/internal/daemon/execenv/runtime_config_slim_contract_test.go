package execenv

import (
	"os"
	"strings"
	"testing"
)

func TestInstalledIssueRuntimeGeneratedContract(t *testing.T) {
	for _, provider := range []string{"claude", "codex"} {
		for _, leader := range []bool{false, true} {
			for _, workflow := range []bool{false, true} {
				t.Run(provider+"/"+map[bool]string{false: "worker", true: "leader"}[leader]+"/"+map[bool]string{false: "plain", true: "workflow"}[workflow], func(t *testing.T) {
					ctx := TaskContextForEnv{IssueID: "issue", IsSquadLeader: leader, HasProjectWorkflow: workflow, AgentSkills: []SkillContextForEnv{issueExecutionFixture("allocated-execution")}, AgentInstructions: "  Role 原文\r\n## Heading\n  ", SquadBriefing: "## Squad Roster\r\n| 人 | agent-id |\n  Instructions  ", ProjectID: "project", ProjectDescription: "  Project 原文\r\n## Description\n  "}
					dir := t.TempDir()
					out, err := InjectRuntimeConfig(dir, provider, ctx)
					if err != nil {
						t.Fatal(err)
					}
					actual, err := os.ReadFile(runtimeConfigPath(dir, provider))
					if err != nil {
						t.Fatal(err)
					}
					if !strings.Contains(string(actual), strings.TrimRight(out, "\n")) {
						t.Fatal("generated provider config lost runtime content")
					}
					out = string(actual) // Validate the actual provider file, including authored bytes and route uniqueness.
					for _, text := range []string{ctx.AgentInstructions, ctx.SquadBriefing, ctx.ProjectDescription} {
						if strings.Count(out, text) != 1 {
							t.Errorf("authored bytes changed or repeated: %q", text)
						}
					}
					for _, text := range []string{"## Definition Boundaries", "MULTICA_CLI_PATH", "Turn exit is terminal", "Never dispatch the same work by both assignment and mention", "--attachment", "Runtime-local paths are never deliverables", "allocated-execution (every issue turn)"} {
						if !strings.Contains(out, text) {
							t.Errorf("missing mechanism %q", text)
						}
					}
					for _, text := range []string{"1. Read the issue", "## Issue Body Formatting", "Gate the cleanup on the post succeeding", "## Instruction Precedence", "## Mentions", "## Attachments"} {
						if strings.Contains(out, text) {
							t.Errorf("method/fallback duplicated: %q", text)
						}
					}
					if strings.Count(out, "allocated-execution (every issue turn)") != 1 {
						t.Fatal("execution routing must have one owner")
					}
					bare := ctx
					bare.AgentInstructions = ""
					bare.SquadBriefing = ""
					bare.ProjectDescription = ""
					if size := len(buildMetaSkillContent(provider, bare)); size > 3072 {
						t.Errorf("platform static body %d bytes exceeds 3072-byte budget", size)
					}
				})
			}
		}
	}
}
