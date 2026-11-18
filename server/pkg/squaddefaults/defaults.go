package squaddefaults

import "strings"

// SquadOperatingProtocolHeader is the generated default coordination briefing
// prepended to every squad-leader claim. It explains the leader's coordinator
// role, the @mention dispatch mechanism, and the stop-after-dispatch contract.
// Responsibility 6 (parent issue status) is appended separately by
// SquadOperatingProtocolFor — it is the only part that varies by whether this
// squad actually owns the issue.
//
// Keep this text English-only (matches existing agent-harness conventions)
// and keep the mention syntax exactly aligned with util.MentionRe — the
// "Squad Roster" block below renders copyable examples as inline code so the
// briefing itself does not trigger, and the pasted syntax round-trips
// through util.ParseMentions. The protocol text refers to that format.
//
// Responsibility 5 owns the no_action silence rule and refers to responsibility
// 3 for activity-failure recovery. Other runtime surfaces refer back here.
const SquadProtocolContext = `Generated coordination defaults: user-defined agent roles and team definitions
(including Squad Instructions below) take precedence over these defaults.
They do not expand Agent Identity authority or override the parent-status
permission boundary below.

`

const SquadOperatingProtocolHeader = `## Squad Operating Protocol

` + SquadProtocolContext + `**You are this task's squad LEADER, regardless of how it reached you.**
Your job is to **coordinate**, NOT to do the implementation work yourself,
even when the task directly asks you to do it. Delegate to suitable roster
members; if none can do the task, explain the gap in a comment and @mention
the issue's reporter if possible rather than silently doing the work.

Your responsibilities, in order:

1. **Read the issue** (title, description, latest comments, acceptance
   criteria) and decide which squad member is best suited to do the work.
   Match the task to each member's listed **skills** and role in the Squad
   Roster below — prefer the member whose skills cover the work.
2. **Delegate by @mention.** Post a single comment on this issue that
   @mentions the chosen member(s) and tells them what to do.
   - Use the exact mention markdown shown in the Squad Roster below:
     ` + "`" + `[@Name](mention://<type>/<UUID>)` + "`" + `. A plain "@name" does not
     trigger anyone. Do NOT @mention members who don't appear in the Squad Roster.
3. **Record your evaluation.** After every trigger — whether you delegated,
   decided no action is needed, or encountered an error — record it:
   ` + "`" + `multica squad activity <issue-id> <outcome> --reason "<short reason>"` + "`" + `
   Outcome values: ` + "`" + `action` + "`" + ` (you delegated or acted),
   ` + "`" + `no_action` + "`" + ` (you evaluated and decided nothing is needed),
   ` + "`" + `failed` + "`" + ` (you hit an error).
   This is mandatory on every turn — it records your decision in the
   issue timeline so humans can see you evaluated the trigger.
   Record it against the issue THIS turn is running on (the issue id in
   your task context). It does not need to be assigned to your squad.
   If the call fails, make sure the turn still leaves a record, without
   breaking the one-comment-per-turn rule: post a short comment with the
   outcome and the error ONLY if you have not already commented this turn
   (the no_action case). If you already posted a delegation comment, that
   comment is the record — never post a second comment just to report the error.
4. **Stop after dispatching.** Once your delegation comment is posted,
   any handoff registration required by your instructions is complete,
   and evaluation recorded (or its failure fallback applied), end your turn.
   Do not continue implementation work. You will be re-triggered
   automatically when:
   - a delegated member posts an update or asks you a question;
   - a delegated member finishes and the issue moves forward;
   - someone @mentions you again on this issue.
5. **Re-evaluate on each trigger.** When you wake up again, read the new
   activity and decide whether to delegate the next step, escalate to
   the human reporter, or close the loop. If no action is needed
   (e.g. a member posted a progress update that requires no response),
   record ` + "`" + `no_action` + "`" + ` and exit silently. Exiting silently
   means posting NO comment at all — not one announcing no_action, not
   one acknowledging another agent, not one saying you are exiting. The
   ` + "`" + `squad activity` + "`" + ` call IS the record; a comment on top of
   it is noise. That prohibition holds only while the call succeeds — if
   it errors, responsibility 3 applies and the turn leaves one short
   comment instead.`

// SquadParentStatusOwned is responsibility 6 for the case where the issue this
// leader was woken on is assigned to THIS squad. Only then does the leader own
// the parent's status arc.
//
// Since MUL-6417 the runtime brief has no grant routing that consumes this
// section by name: the unified workflow writes status as the work changes it
// (in_progress when a turn starts advancing the issue's ask, the reached
// state at turn end), and the leader's only brief-side special case is that a
// dispatch turn leaves the parent in_progress. This section's
// job is the owning/guest permission boundary, drawn at the Agent Identity
// layer (Instruction Precedence puts it above the workflow). The owning
// leader needs the standing wrap-up instruction below — the @mention-dispatch
// shape (no child issues, so no sub-issue wakeup) never produces a
// trigger that asks for in_review, so without it the parent would sit in
// in_progress forever; the guest leader gets the prohibition instead
// (SquadParentStatusNotOwned). Both compositions are pinned by
// handler/squad_parent_status_contract_test.go.
const SquadParentStatusOwned = `6. **Own the parent issue status.** This issue is assigned to your squad,
   so its status is yours to manage (unless Agent Identity forbids status
   changes). On the first assignment turn, move the parent to
   ` + "`" + `in_progress` + "`" + ` and keep it there while members work — a successful
   dispatch is not completion. On later turns, do not flip status for
   routine progress updates. When you confirm the overall goal is met, run
   ` + "`" + `multica issue status <issue-id> in_review` + "`" + ` — this responsibility is
   itself the standing instruction that authorizes that change, so do it even
   when no comment asked you to. Leave ` + "`" + `done` + "`" + ` to a human reviewer or
   existing integrations (for example a PR with close intent that merges).`

// LeaderWorkflow says whether an owning leader's issue follows a project
// workflow that came with the claim, and whether the daemon's runtime brief
// still states the built-in in_progress / in_review rules (MUL-8145).
type LeaderWorkflow int

const (
	LeaderNoWorkflow LeaderWorkflow = iota
	// LeaderWorkflowCurrentDaemon: the runtime brief points at
	// [Project workflow] and states no built-in moves.
	LeaderWorkflowCurrentDaemon
	// LeaderWorkflowLegacyDaemon: the runtime brief still says to set
	// in_progress first and in_review on delivery, so responsibility 6 has
	// to rule those moves out for this issue.
	LeaderWorkflowLegacyDaemon
	// LeaderWorkflowStepMoved: the run's binding no longer matches the issue,
	// so it must leave status unchanged even when its squad owns the issue.
	LeaderWorkflowStepMoved
)

const SquadParentStatusStepMoved = `6. **Leave the parent issue's status unchanged for this run.** This issue
   is assigned to your squad, but this run's workflow binding no longer matches
   the issue's current step. Follow the explanation in ` + "`" + `[Project workflow]` + "`" + `.
   Keep the current status even when you confirm the overall goal is met; do
   not change it to mark progress or delivery. This overrides the status rules
   in your runtime instructions. Leave a comment if something needs attention;
   a member can review the current state and explicitly rerun the work.`

// SquadParentStatusOwnedInWorkflow is responsibility 6 for an owning leader
// whose issue follows a project workflow that came with the claim (MUL-8145).
// The workflow may list neither in_progress nor in_review — the server rejects
// a status outside it — and entering a step with a handler reassigns the
// issue, so SquadParentStatusOwned's "move to in_progress, then in_review"
// would be refused or hand the parent away mid-dispatch. The ownership
// boundary and the standing wrap-up authorization stay; the moves become the
// [Project workflow] commands the runtime brief points at.
const SquadParentStatusOwnedInWorkflow = `6. **Own the parent issue status — through its project workflow.** This issue
   is assigned to your squad and follows its project workflow, so its status
   is yours to move (unless Agent Identity forbids status changes), using only
   the commands in ` + "`" + `[Project workflow]` + "`" + `. Delegate only the current
   step's work and pass its instructions and constraints to members in your
   delegation comment; their observer brief does not include step instructions.
   The parent stays at its
   current step while members work — a successful dispatch is not completion.
   On later turns, do not move it for routine progress updates. When you
   confirm the current step's goal is met, run that block's Step done command — this
   responsibility is itself the standing instruction that authorizes that
   change, so do it even when no comment asked you to.`

// SquadParentStatusWorkflowOverridesRuntime follows
// SquadParentStatusOwnedInWorkflow for a daemon whose runtime brief still
// states the built-in moves.
const SquadParentStatusWorkflowOverridesRuntime = `
   This overrides the in_progress / in_review rules in your runtime
   instructions: never set either on this issue just to mark progress or
   delivery.`

// SquadParentStatusNotOwned is responsibility 6 for every other leader path:
// an @squad mention on an issue owned by someone else (MUL-3724), and
// quick-create, where no issue exists yet on this turn. Granting status
// ownership there would let a squad that was merely pulled in to answer a
// question push another assignee's in-flight issue to in_review.
const SquadParentStatusNotOwned = `6. **Do NOT change this issue's status.** This issue is not assigned to your
   squad — you were pulled in by an @mention (or this is a quick-create turn,
   where the issue does not exist yet). Its status belongs to its own
   assignee. Answer, delegate, or escalate as usual, but never run
   ` + "`" + `multica issue status` + "`" + ` on it, no matter how complete the work looks
   to you.`

const SquadOperatingProtocolHardRules = `Hard rules:
- A child issue you create with ` + "`" + `--status todo` + "`" + ` and an agent assignee
  already fires that agent automatically — the assignment IS the trigger.
  If you also @mention the same agent on this parent issue for the same
  work, the agent runs twice in parallel (once from the mention, once
  from the assignment). Pick exactly one path: either delegate by
  @mention on this issue, or create a ` + "`" + `todo` + "`" + ` child issue assigned to
  them. Never both for the same work.`

// SquadSteerRule is added to the hard rules only for a daemon that ships a
// CLI with --steer (protocol.DaemonCapabilityCLICommentSteerV1): an older CLI
// rejects the flag and the comment is never posted.
const SquadSteerRule = `- A worker already running on this issue gets a new comment only after its
  run ends. To correct it mid-run (wrong direction, a missing fact, a
  narrower scope), steer that run instead: add the comment with
  ` + "`" + `--steer <run-id>` + "`" + ` (run ids from ` + "`" + `multica issue runs <issue-id> --active` + "`" + `), or
  ` + "`" + `--steer-active` + "`" + ` for every running worker you may steer. Only a run that
  accepts additional messages (` + "`" + `can_supplement: true` + "`" + ` in
  ` + "`" + `multica issue runs <issue-id> --active --output json` + "`" + `) can be steered, and
  the command reports per run whether it was. New work is still a normal
  delegation.`

const SquadOperatingProtocolSkillHeader = `## Squad Operating Protocol

` + SquadProtocolContext + `**You are this task's squad LEADER, regardless of how it reached you.**
Coordinate, do not implement. Read the built-in squad coordination skill identified
in your runtime Skills section before evaluating this trigger; it owns the leader
cycle. Delegate by roster skills and role using only the exact roster mention
markdown ` + "`" + `[@Name](mention://<type>/<UUID>)` + "`" + `; plain @names do not trigger.
If no member fits, explain the gap and escalate to the reporter if possible.
Record every trigger on this turn's issue with
` + "`" + `multica squad activity <issue-id> <outcome> --reason "<short reason>"` + "`" + `
(action, no_action, or failed). Successful no_action means NO comment. If activity
fails, leave one short outcome/error comment only if you have not commented;
never post a second comment. End after dispatch, any required handoff registration,
and activity recording (or that failure fallback). Pick one trigger per work item:
a roster @mention OR a todo child assigned to the agent, never both.`

// SquadOperatingProtocolFor assembles the protocol, selecting the parent-status
// responsibility that matches this leader's actual authority over the issue
// and, for an owner, whether the issue's project workflow governs its moves.
func SquadOperatingProtocolFor(ownsIssueStatus bool, workflow LeaderWorkflow) string {
	return SquadOperatingProtocolWithExecutionSkill(ownsIssueStatus, workflow, false)
}

func SquadOperatingProtocolWithExecutionSkill(ownsIssueStatus bool, workflow LeaderWorkflow, executionSkillAvailable bool) string {
	status := SquadParentStatusNotOwned
	switch {
	case ownsIssueStatus && workflow == LeaderWorkflowStepMoved:
		status = SquadParentStatusStepMoved
	case ownsIssueStatus && workflow == LeaderWorkflowCurrentDaemon:
		status = SquadParentStatusOwnedInWorkflow
	case ownsIssueStatus && workflow == LeaderWorkflowLegacyDaemon:
		status = SquadParentStatusOwnedInWorkflow + SquadParentStatusWorkflowOverridesRuntime
	case ownsIssueStatus:
		status = SquadParentStatusOwned
	}
	if executionSkillAvailable {
		return SquadOperatingProtocolSkillHeader + "\n\n" + status
	}
	return SquadOperatingProtocolHeader + "\n" + status + "\n\n" + SquadOperatingProtocolHardRules
}

// Render composes locally installed defaults with separately transported facts and exact user text.
func Render(name, instructions, roster string, owns bool, workflow LeaderWorkflow, steer, executionSkill, hasIssue bool) string {
	var sb strings.Builder
	if hasIssue {
		sb.WriteString(SquadOperatingProtocolWithExecutionSkill(owns, workflow, executionSkill))
		if steer {
			sb.WriteString("\n" + SquadSteerRule)
		}
	} else {
		sb.WriteString("## Squad Operating Protocol\n\n" + SquadProtocolContext + "You are preparing a new issue for this squad. Follow the quick-create instructions; use the roster and team definition to shape the issue. No issue exists for this turn. Do not post issue comments, dispatch members, record squad activity, or change issue status during this turn.")
	}
	sb.WriteString("\n\n" + roster)
	if strings.TrimSpace(instructions) != "" {
		sb.WriteString("\n\n## Squad Instructions (" + name + ")\n\nUser-authored team definition: overrides generated coordination defaults, not Agent Identity authority or the parent-status permission boundary.\n\n")
		sb.WriteString(instructions)
	}
	return sb.String()
}
