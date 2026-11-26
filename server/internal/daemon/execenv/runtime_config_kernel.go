package execenv

import (
	"fmt"
	"path/filepath"
	"strings"
)

// buildRuntimeKernel contains execution mechanisms, not Skill procedures.
func buildRuntimeKernel(provider string, ctx TaskContextForEnv) string {
	var b strings.Builder
	kind := classifyTask(ctx)
	fmt.Fprintf(&b, "# Multica Runtime\n\nUse only the daemon CLI: %s (MULTICA_CLI_PATH), never PATH, curl or wget. Discover syntax with --help; JSON stdout and diagnostic stderr are separate. A parse error is not permission to repeat a successful write.\n\n", multicaCLICommand())
	b.WriteString("## Definition Boundaries\n\nRuntime permissions, task kind and injected status authority are binding. Agent Identity sets role/prohibitions; Squad process and this turn's goal stay within them. User definitions override built-in defaults, not runtime permissions. Skills never grant extra authorization; skip conflicting actions, including implementation in a delegation-only role or forbidden status writes.\n\n")
	b.WriteString("## Run Lifecycle\n\nTurn exit is terminal for this Run: collect required run-owned results and deliver before exiting; never background-and-yield. There is no same-Run background-completion wakeup. Issue wakeups start a new Run. Do not terminate the Multica daemon that carries this and sibling Runs.\n\n")
	b.WriteString("## Dispatch Boundaries\n\nAgent assignment can start a run; an agent-assigned todo child starts immediately, backlog parks it. Human mentions notify; agent mentions enqueue runs. Never dispatch the same work by both assignment and mention or repeat an existing handoff. Courtesy/FYI is not new work.\n\n")
	b.WriteString("## Skills\n\nBefore acting, load the installed contract(s) for this surface and follow their reference routing: ")
	skills := modelVisibleSkills(ctx.AgentSkills)
	var routes []string
	if slug, ok := builtinSlug(skills, platformSkillName); ok {
		routes = append(routes, slug+" (platform actions)")
	}
	if kind == kindIssue {
		if slug, ok := builtinSlug(skills, issueExecutionSkillName); ok {
			routes = append(routes, slug+" (every issue turn)")
		}
		if ctx.IsSquadLeader {
			if slug, ok := builtinSlug(skills, "multica-squad-coordination"); ok {
				routes = append(routes, slug+" (leader coordination)")
			}
		}
	}
	for i, route := range routes {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(route)
	}
	if provider == "traecli" {
		fmt.Fprintf(&b, ". Read each from %s/<allocated-name>/SKILL.md", filepath.ToSlash(skillsDirPath("", provider)))
	}
	b.WriteString(". Skill availability does not authorize mutations.\n\n")
	b.WriteString("## Output\n\nRuntime-local paths are never deliverables or clickable links. Use this surface's delivery mechanism; report a gap rather than implying a local file was delivered.\n")
	switch kind {
	case kindIssue:
		b.WriteString("The per-turn message supplies issue identity, trigger, reply threads and status authority; never reuse old values. ")
		if ctx.HasProjectWorkflow {
			b.WriteString("[Project workflow] alone defines status commands. ")
		}
		b.WriteString("Terminal output and run logs are not user delivery. Post only final results before exit via multica issue comment add: one comment, or one reply per listed thread; files require --attachment. A supplied WAKEUP checkin may replace an empty check. ")
		if ctx.IsSquadLeader {
			b.WriteString("After successfully recording no_action activity for this turn, the leader must stay silent; if recording fails, follow the coordination Skill fallback. action/failed require a comment. ")
		}
	case kindChat:
		b.WriteString("Reply text is delivered directly to chat. ")
		if ctx.ChatChannelType != "" {
			b.WriteString("File delivery depends on deployment: follow this turn's delivery instructions. ")
		} else {
			b.WriteString("Files reach chat only through multica attachment upload. ")
		}
	case kindAutopilotRunOnly:
		b.WriteString("Run-only autopilot has no assigned issue. Final output is captured as its text-only run result. Issue reads/comments/status require explicit autopilot instructions. ")
	case kindQuickCreate:
		b.WriteString("Quick-create has NO existing issue. Invoke issue create --output json exactly once, then exit; never retry, even on error. Do not issue get, status or comment. On success print only Created <identifier-or-id>: <title> from returned JSON; on CLI/parse error print only that error. Platform notifications are automatic. Stdout is text-only; issue files belong on create via --attachment. ")
	}
	b.WriteString("\n\n")
	writeRuntimeSkillMetadata(&b, provider, skills)
	// Authored context and transport facts are not platform static rules.
	writeRequestingUser(&b, ctx)
	writeWorkspaceContext(&b, ctx)
	if ctx.ProjectID != "" || len(ctx.ProjectResources) > 0 {
		b.WriteString("## Project Context\n\n")
		if ctx.ProjectTitle != "" {
			fmt.Fprintf(&b, "Project: %s\n\n", ctx.ProjectTitle)
		}
		if ctx.ProjectDescription != "" {
			b.WriteString(ctx.ProjectDescription)
			b.WriteString("\n\n")
		}
		for _, r := range ctx.ProjectResources {
			fmt.Fprintf(&b, "- %s\n", formatProjectResource(r))
		}
		b.WriteString("\n")
	}
	if len(ctx.Repos) > 0 {
		b.WriteString("## Repositories\n\n")
		for _, repo := range ctx.Repos {
			fmt.Fprintf(&b, "- %s", repo.URL)
			if repo.Description != "" {
				fmt.Fprintf(&b, " — %s", repo.Description)
			}
			if repo.Ref != "" {
				fmt.Fprintf(&b, " (starts from %s)", repo.Ref)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if len(ctx.IssueStatuses) > 0 {
		writeIssueStatusCommand(&b, ctx)
	}
	if ctx.AgentName != "" || ctx.AgentID != "" || ctx.AgentInstructions != "" {
		name := ctx.AgentName
		if name == "" {
			name = "Agent"
		}
		fmt.Fprintf(&b, "# %s Definition\n\n", sanitizeNameForBriefMarkdown(name))
		writeAgentIdentity(&b, ctx)
	}
	if ctx.SquadBriefing != "" {
		b.WriteString("# Team Information\n\n")
		b.WriteString(ctx.SquadBriefing)
		b.WriteString("\n\n")
	}
	return b.String()
}

// writeRuntimeSkillMetadata preserves the native names index and file-only discovery.
func writeRuntimeSkillMetadata(b *strings.Builder, provider string, skills []SkillContextForEnv) {
	b.WriteString("<!-- BEGIN MULTICA-SKILL-METADATA -->\n## Available Skills\n\n")
	for _, skill := range skills {
		fmt.Fprintf(b, "- **%s**", skill.Name)
		if provider == "traecli" {
			description := []rune(strings.Join(strings.Fields(skill.Description), " "))
			if len(description) > 240 {
				description = append(description[:239], '…')
			}
			if len(description) > 0 {
				fmt.Fprintf(b, " — %s", sanitizeNameForBriefMarkdown(string(description)))
			}
			fmt.Fprintf(b, " (file: %s)", filepath.ToSlash(filepath.Join(skillsDirPath("", provider), skill.Name, "SKILL.md")))
		}
		b.WriteString("\n")
	}
	b.WriteString("<!-- END MULTICA-SKILL-METADATA -->\n\n")
}
