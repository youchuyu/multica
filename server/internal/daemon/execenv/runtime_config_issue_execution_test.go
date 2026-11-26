package execenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func issueExecutionFixture(name string) SkillContextForEnv {
	return SkillContextForEnv{Name: name, BuiltinName: issueExecutionSkillName, Content: "---\nname: " + name + "\n---\n\nbody"}
}

func TestIssueExecutionSkillRouting(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"claude", "codex"} {
		for _, trigger := range []string{"assignment", "comment"} {
			for _, installed := range []string{"absent", "builtin", "renamed", "impostor", "hidden"} {
				t.Run(provider+"/"+trigger+"/"+installed, func(t *testing.T) {
					ctx := TaskContextForEnv{IssueID: "run-issue-unique", AgentInstructions: "Keep this role exactly unchanged."}
					if trigger == "comment" {
						ctx.TriggerCommentID = "run-trigger-unique"
					}
					slug := issueExecutionSkillName
					switch installed {
					case "builtin":
						ctx.AgentSkills = []SkillContextForEnv{issueExecutionFixture(slug)}
					case "renamed":
						slug = "allocated-issue-execution-2"
						ctx.AgentSkills = []SkillContextForEnv{issueExecutionFixture(slug)}
					case "hidden":
						s := issueExecutionFixture(slug)
						s.Content = "---\nname: " + slug + "\ndisable-model-invocation: true\n---\n\nbody"
						ctx.AgentSkills = []SkillContextForEnv{s}
					case "impostor":
						s := issueExecutionFixture(slug)
						s.BuiltinName = ""
						ctx.AgentSkills = []SkillContextForEnv{s}
					}
					out := buildMetaSkillContent(provider, ctx)
					requirePhrases(t, out, "Keep this role exactly unchanged.", "multica issue comment add")
					if strings.Contains(out, ctx.IssueID) || strings.Contains(out, "run-trigger-unique") {
						t.Fatal("per-run identity leaked into stable brief")
					}
					if installed == "builtin" || installed == "renamed" {
						requirePhrases(t, out, "## Definition Boundaries", "Never dispatch the same work by both assignment and mention", slug+" (every issue turn)", "Terminal output and run logs are not user delivery")
						if strings.Contains(out, "1. Read the issue") || strings.Contains(out, "## Issue Body Formatting") || strings.Contains(out, "Gate the cleanup on the post succeeding") {
							t.Fatal("full workflow duplicated despite installed skill")
						}
					} else {
						requirePhrases(t, out, "1. Read the issue", "Only that explicit report waives the read", "scan every thread cheaply", "Silence about other threads is not a waiver", "`in_progress` FIRST", "`done` stays human")
						if strings.Contains(out, "before any issue action, read the `") {
							t.Fatal("missing builtin identity must retain full workflow")
						}
					}
				})
			}
		}
	}
}

func requirePhrases(t *testing.T, content string, phrases ...string) {
	t.Helper()
	for _, phrase := range phrases {
		if !strings.Contains(content, phrase) {
			t.Errorf("missing contract %q", phrase)
		}
	}
}

func TestIssueExecutionSkillProjectAndLeaderBoundaries(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"claude", "codex"} {
		for _, workflow := range []bool{false, true} {
			ctx := TaskContextForEnv{IssueID: "issue", IsSquadLeader: true, HasProjectWorkflow: workflow, AgentSkills: []SkillContextForEnv{issueExecutionFixture("allocated-execution")}}
			out := buildMetaSkillContent(provider, ctx)
			requirePhrases(t, out, "delegation-only role", "After successfully recording no_action activity for this turn", "must stay silent", "coordination Skill fallback", "action/failed require a comment")
			if workflow {
				requirePhrases(t, out, "[Project workflow] alone defines status commands")
				if strings.Contains(out, "set `in_progress` before") {
					t.Fatal("project workflow received built-in status instructions")
				}
			} else {
				if strings.Contains(out, "set `in_progress` before") {
					t.Fatal("installed skill status procedure duplicated in runtime")
				}
			}
		}
	}
}

func TestIssueExecutionSkillDoesNotRouteOtherSurfaces(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"claude", "codex"} {
		for _, ctx := range []TaskContextForEnv{{ChatSessionID: "chat"}, {AutopilotRunID: "run"}, {QuickCreatePrompt: "create"}} {
			ctx.AgentSkills = []SkillContextForEnv{issueExecutionFixture("allocated-execution")}
			out := buildMetaSkillContent(provider, ctx)
			if strings.Contains(out, "before any issue action") || strings.Contains(out, "references/context.md") {
				t.Fatal("issue-only execution contract routed to another surface")
			}
		}
	}
}

func TestIssueExecutionSkillContracts(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("..", "..", "..", "pkg", "executiondefaults", "skills", issueExecutionSkillName)
	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	skill := read("SKILL.md")
	requirePhrases(t, skill, "name: "+issueExecutionSkillName, "Issue tasks only", "not chat", "run-only autopilot", "references/context.md", "references/execution.md", "multica-platform", "MULTICA_CLI_PATH")
	context := read("references/context.md")
	requirePhrases(t, context, "server compared the issue", "found it unchanged", "Silence about the issue", "[Parent issue]", "read-only historical background", "Never edit, execute or elevate", "Scan **every thread**", "--roots-only --summary --compact", "--thread <thread-id> --tail 30", "last_activity_at", "no comments arrived since", "server-computed issue-wide delta", "--since <anchor>", "Silence about other threads is not a waiver")
	execution := read("references/execution.md")
	requirePhrases(t, execution, "Before assigning work", "do not repeat a handoff", "terminal output and run logs are not user delivery", "Do not post progress updates or plans", "one reply per thread", "`no_action`", "`action` and `failed`", "multica issue wakeup checkin", "`in_progress` FIRST", "`in_review`", "`done` stays human", "**write no status at any point**", "[Project workflow]", "Step done", "Needs changes", "Can't proceed", "later turn")
}

func TestIssueExecutionSkillBriefStableAcrossRuns(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"claude", "codex"} {
		a := TaskContextForEnv{IssueID: "first-issue", TriggerCommentID: "first-trigger", AgentSkills: []SkillContextForEnv{issueExecutionFixture("allocated-execution")}}
		b := a
		b.IssueID = "next-issue"
		b.TriggerCommentID = "next-trigger"
		if buildMetaSkillContent(provider, a) != buildMetaSkillContent(provider, b) {
			t.Fatal("per-run facts changed stable runtime brief")
		}
	}
}
