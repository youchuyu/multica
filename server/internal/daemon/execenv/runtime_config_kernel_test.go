package execenv

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeKernelSkillDiscovery(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "traecli"} {
		platform := platformSkillFixture()
		platform.Content += "\n" + platformOperationContractMarker
		hidden := issueExecutionFixture("hidden")
		hidden.Content = "---\nname: hidden\ndisable-model-invocation: true\n---\nbody"
		ctx := TaskContextForEnv{ChatSessionID: "chat", AgentSkills: []SkillContextForEnv{platform, {Name: "user-review", Description: "Review user changes", Content: "---\nname: user-review\n---\nbody"}, hidden}}
		out := buildMetaSkillContent(provider, ctx)
		requirePhrases(t, out, "**user-review**", "**"+platform.Name+"**")
		if strings.Contains(out, "**hidden**") {
			t.Fatal("hidden skill exposed")
		}
		if provider == "traecli" {
			requirePhrases(t, out, "Review user changes", filepath.ToSlash(filepath.Join(skillsDirPath("", provider), "user-review", "SKILL.md")))
		}
		if strings.Count(out, "(platform actions)") != 1 {
			t.Fatal("platform route duplicated")
		}
	}
}

func TestRuntimeKernelKeepsOnlyPlatformLifecycle(t *testing.T) {
	ctx := TaskContextForEnv{IssueID: "issue", IsSquadLeader: true, AgentSkills: []SkillContextForEnv{issueExecutionFixture("execution")}}
	out := buildMetaSkillContent("claude", ctx)
	requirePhrases(t, out, "collect required run-owned results and deliver before exiting", "Issue wakeups start a new Run", "After successfully recording no_action activity for this turn", "must stay silent", "coordination Skill fallback", "action/failed require a comment")
	for _, general := range []string{"external CI", "foreground blocking call", "durable logs", "cleanup handle", "unsupervised survival", "gh pr create"} {
		if strings.Contains(out, general) {
			t.Errorf("general development advice injected: %q", general)
		}
	}
	leaderReply := BuildCommentReplyInstructions("claude", "issue", "comment", true)
	requirePhrases(t, leaderReply, "successfully recorded `no_action` activity", "--parent comment", "platform Skill issues reference")
	fanout := BuildMultiThreadCommentReplyInstructions("issue", []ThreadReplyTarget{{ThreadID: "t1", ParentID: "p1"}, {ThreadID: "t2", ParentID: "p2"}}, true)
	requirePhrases(t, fanout, "successfully recorded `no_action` activity", "If activity recording failed", "only if no comment was already posted", "--parent p1", "--parent p2")
	if strings.Contains(fanout, "Comment Formatting` above") {
		t.Fatal("fan-out references removed runtime section")
	}
	for _, hint := range []string{BuildNewCommentsHint("issue", "comment", "thread", "2026-11-19T00:00:00Z", 2), BuildResumedCommentsHint("issue", "comment", "thread"), BuildResumedUnknownDeltaCommentsHint("issue", "comment", "thread"), BuildColdCommentsHint("issue", "comment", "thread"), BuildIssueStateHint("issue", "todo", "agent", "agent", nil, true, true)} {
		if strings.Contains(hint, "workflow step") {
			t.Errorf("dangling numbered workflow reference: %s", hint)
		}
	}
}

func TestRuntimeKernelMissingIssueContractFallback(t *testing.T) {
	for _, state := range []string{"absent", "hidden", "impostor"} {
		platform := platformSkillFixture()
		platform.Content += "\n" + platformOperationContractMarker
		ctx := TaskContextForEnv{IssueID: "issue", AgentSkills: []SkillContextForEnv{platform}}
		s := issueExecutionFixture("execution")
		if state == "hidden" {
			s.Content = "---\nname: execution\ndisable-model-invocation: true\n---\nbody"
		}
		if state == "impostor" {
			s.BuiltinName = ""
		}
		if state != "absent" {
			ctx.AgentSkills = append(ctx.AgentSkills, s)
		}
		for _, provider := range []string{"claude", "codex"} {
			out := buildMetaSkillContent(provider, ctx)
			requirePhrases(t, out, "1. Read the issue", "Only that explicit report waives the read")
			if strings.Contains(out, "(every issue turn)") {
				t.Fatal("unavailable issue contract routed")
			}
		}
	}
}
