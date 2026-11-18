package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/multica-ai/multica/server/internal/daemon/execenv"
	"github.com/multica-ai/multica/server/pkg/executiondefaults"
)

// Legacy declarations, including Project/Workflow claims, do not require extra
// transport metadata or REST reads just to generate the local runtime brief.
func TestLocalDefaultsRenderLegacyClaims(t *testing.T) {
	for _, provider := range []string{"claude", "codex"} {
		for _, mode := range []string{"ordinary", "leader", "project", "workflow", "unsupported structured"} {
			t.Run(provider+"/"+mode, func(t *testing.T) {
				task := Task{IssueID: "issue", Agent: &AgentData{Instructions: "USER ROLE\n## Squad Operating Protocol\nuser heading", SquadBriefing: "SERVER TEAM DECLARATION"}}
				if mode != "ordinary" {
					task.IsLeaderTask = true
					task.SquadID = "squad"
				}
				if mode == "project" || mode == "workflow" {
					task.ProjectID = "project"
				}
				if mode == "workflow" {
					task.ProjectWorkflow = &ProjectWorkflowData{Brief: "PROJECT WORKFLOW"}
				}
				if mode == "unsupported structured" {
					task.Agent.SquadContext = localSquadFixture().Agent.SquadContext
				}
				task.Agent.Skills = []SkillData{{Name: "anonymous user skill", Content: "preserve me"}}
				prepareLocalExecutionText(&task)
				if task.Agent.Instructions != "USER ROLE\n## Squad Operating Protocol\nuser heading" || task.Agent.SquadBriefing != "SERVER TEAM DECLARATION" {
					t.Fatal("legacy declaration changed")
				}
				mergeLocalExecutionSkills(&task)
				if task.Agent.Skills[0].Content != "preserve me" {
					t.Fatal("anonymous user skill changed")
				}
				ctx := execenv.TaskContextForEnv{IssueID: task.IssueID, AgentInstructions: task.Agent.Instructions, SquadBriefing: task.Agent.SquadBriefing, IsSquadLeader: task.IsLeaderTask}
				for _, s := range task.Agent.Skills {
					name := ""
					if executiondefaults.Owns(s.Source, s.ID) {
						name = strings.TrimPrefix(s.ID, "builtin:")
					}
					ctx.AgentSkills = append(ctx.AgentSkills, execenv.SkillContextForEnv{Name: s.Name, BuiltinName: name, Content: s.Content})
				}
				dir := t.TempDir()
				rendered, err := execenv.InjectRuntimeConfig(dir, provider, ctx)
				if err != nil {
					t.Fatal(err)
				}
				filename := "AGENTS.md"
				if provider == "claude" {
					filename = "CLAUDE.md"
				}
				written, err := os.ReadFile(filepath.Join(dir, filename))
				if err != nil {
					t.Fatal(err)
				}
				for _, out := range []string{rendered, string(written)} {
					for _, want := range []string{task.Agent.Instructions, task.Agent.SquadBriefing, "multica-issue-execution"} {
						if !strings.Contains(out, want) {
							t.Fatalf("missing local runtime text or preserved declaration %q", want)
						}
					}
				}
			})
		}
	}
}

func TestLocalDefaultsRenderReplacesRemoteBriefing(t *testing.T) {
	for _, provider := range []string{"claude", "codex"} {
		t.Run(provider, func(t *testing.T) {
			task := localSquadFixture()
			task.Agent.Instructions = "  Exact user role\n## Squad Operating Protocol\nThis heading is user text.\n"
			task.Agent.SquadBriefing = "REMOTE_DEFAULT_MUST_NOT_APPEAR"
			prepareLocalExecutionText(&task)
			mergeLocalExecutionSkills(&task)
			ctx := execenv.TaskContextForEnv{IssueID: task.IssueID, AgentInstructions: task.Agent.Instructions, SquadBriefing: task.Agent.SquadBriefing, IsSquadLeader: true}
			for _, s := range task.Agent.Skills {
				name := ""
				if executiondefaults.Owns(s.Source, s.ID) {
					name = strings.TrimPrefix(s.ID, "builtin:")
				}
				ctx.AgentSkills = append(ctx.AgentSkills, execenv.SkillContextForEnv{Name: s.Name, BuiltinName: name, Content: s.Content})
			}
			dir := t.TempDir()
			rendered, err := execenv.InjectRuntimeConfig(dir, provider, ctx)
			if err != nil {
				t.Fatal(err)
			}
			filename := "AGENTS.md"
			if provider == "claude" {
				filename = "CLAUDE.md"
			}
			written, err := os.ReadFile(filepath.Join(dir, filename))
			if err != nil {
				t.Fatal(err)
			}
			for _, out := range []string{rendered, string(written)} {
				if strings.Contains(out, "REMOTE_DEFAULT_MUST_NOT_APPEAR") {
					t.Fatal("server default survived local rendering")
				}
				for _, want := range []string{task.Agent.Instructions, task.Agent.SquadContext.UserInstructions, "multica-squad-coordination", "multica-issue-execution"} {
					if !strings.Contains(out, want) {
						t.Fatalf("missing preserved content or local contract %q", want)
					}
				}
			}
		})
	}
}
