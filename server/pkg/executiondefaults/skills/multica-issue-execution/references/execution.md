# Execution, Delivery and Status

After the mandatory context reads, complete the task within Agent Identity's
boundaries. Skip conflicting actions, not the remaining compatible steps. A
delegation-only role performs permitted delegation and stops when it is delivered.
Before assigning work, check the target issue's current assignee and comment
history: do not repeat a handoff already made or being handled by that assignee.

## Real delivery

Post only final results before this turn exits, through
`multica issue comment add`: terminal output and run logs are not user delivery.
Do not post progress updates or plans. With a triggering comment, use **this
turn's** supplied `--parent`, never a previous turn's; with several listed
threads, post one reply per thread. Without a trigger, post a top-level comment.
Read the multica-platform issues reference for the detailed file-first comment,
receipt, attachment, PR-link and chart contracts; do not invent delivery mechanics.

For squad leaders, the coordination skill owns activity and silence: only a valid
`no_action` outcome waives the final comment; `action` and `failed` require one.
Only a scheduled wakeup check whose `[WAKEUP]` block offers
`multica issue wakeup checkin` may check in instead, when nothing needs a reply.
Follow the runtime's Output rules, including never linking runtime-local paths.

## Status without a project workflow

Status reflects the **issue's deliverable**, not the run's lifecycle or the kind
of activity (research, design, planning, review). Skip any status call forbidden
by Agent Identity. Write only a changed value, whoever the assignee is, as soon
as work changes the state, including mid-turn:

- Before producing any part of this issue's ask, set `in_progress` FIRST unless
  already there. Do not wait until the work is finished to show it being worked.
- Delivering what the issue itself asks for, awaiting acceptance, goes to
  `in_review`, including an assigned sub-issue in a chain or stage. This is ready
  for acceptance, not a closed stage barrier. `done` stays human; stage closure
  and parent wakeup conditions follow the platform sub-issues reference.
- Work continuing beyond this turn (dispatched children, one part delivered with
  more underway) stays `in_progress`.
- Unable to proceed without something missing goes to `blocked`; explain the
  blocker in a comment unless Agent Identity forbids commenting.
- Questions, discussion, acknowledgements or consulting on work owned elsewhere
  produce none of this issue's own deliverable: **write no status at any point**.
  This no-write default prevents concurrent runs from flapping the board.

These rules name exact built-in keys. Custom statuses share lifecycle semantics
only, not built-in automation behavior.

A squad leader dispatching members has not delivered the parent. Leave it
`in_progress`; move it to `in_review` only on a **later** member-update or
stage-barrier turn that confirms the overall goal is met.

## Status with a project workflow

A per-turn `[Project workflow]` is this issue's status contract: its statuses,
current step and commands override built-in status assumptions. Use **only its
commands**; do not set `in_progress` or `in_review` merely because the generic
workflow names them. A workflow may not contain those keys, and entering a
handoff step reassigns the issue and starts that handler's run.

- Step finished: run **Step done**.
- Rework needed at an earlier step: run **Needs changes**, when available.
- Work continues beyond this turn: leave the current step unchanged.
- Blocked: follow **Can't proceed**, explaining the blocker unless comments are
  forbidden by Agent Identity.
- No issue deliverable produced: write nothing, at any point.

Squad dispatch leaves the parent at its current step; run Step done only on the
later turn that confirms the current step's goal is met. Before exiting, confirm that
you ran the command matching the actual outcome, or deliberately left the issue
at its current step because work continues. Without a project workflow, confirm
that its status likewise still matches reality.
