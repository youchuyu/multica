---
name: multica-squad-coordination
description: "Use when the runtime identifies you as a squad leader: evaluate a trigger, route work to roster members, record squad activity, and close the coordination loop. Not for general issue execution or product implementation."
user-invocable: false
allowed-tools: Bash(multica *)
---

# Squad coordination

Use this skill only when the runtime identifies
you as the squad leader. Coordinate; do not implement the task yourself, even
when its wording asks you to. Your runtime brief owns identity, instruction
precedence, comment policy, and parent-status authority. Follow its owning,
guest, or project-workflow status rules exactly; this skill grants no additional
status authority and does not replace the general issue workflow. Follow the
user Agent/Squad definitions when their process differs from these defaults.

This cycle applies only to turns on an existing issue. Quick-create has no
current issue: follow its create-once workflow instead, without issue reads,
delegation comments or squad activity calls. A roster or leader flag alone does
not authorize those operations.

## Each trigger

1. Read the current issue, acceptance criteria, trigger and relevant new activity.
   Use the issue-execution skill's bounded comment reads. Check what members already did or
   are doing before dispatching duplicate work.
2. Decide whether to delegate a next step, answer a coordination question,
   escalate a blocker to the human reporter, or close the loop. Match work to
   the skills and roles in the Squad Roster. If no member fits, explain the gap
   and mention the reporter if possible; do not silently do the work yourself.
3. Choose one dispatch path: a roster mention on this issue, or an assigned
   `todo` child as described below. For the mention path, post a single comment
   with the chosen member's exact roster mention markdown:
   `[@Name](mention://<type>/<UUID>)`. Plain `@name` does not trigger anyone;
   never invent members or mention links. Give the member an actionable task
   under the runtime comment policy. For a project-workflow step, delegate only
   that step's work and include its instructions and constraints: members'
   observer briefs do not include them.
4. Complete any handoff registration required by your installed instructions
   before stopping. Do not assume the delegation comment alone satisfies it.
   No particular handoff skill or registration mechanism is required by this
   skill; use one only when the task's actual instructions require it.
5. Record the evaluation on the issue **this turn runs on**, even when it is
   not assigned to your squad:

   `multica squad activity <issue-id> <outcome> --reason "<short reason>"`

   Use `action` when you delegated or acted, `no_action` when evaluation found
   nothing needed, or `failed` when an error prevented the intended action.
   Every trigger requires this record, including errors and routine updates.
6. End the turn after dispatch, required handoff registration, and activity
   recording or its failure fallback. Do not continue implementation. Member
   updates, completion, questions, or a new mention will trigger reevaluation.
   On later turns, assess the new facts rather than acknowledging every update.

## Silence and errors

Successful `no_action` means **no comment at all**: no acknowledgement, no
announcement of no_action, and no message saying you are exiting. The activity
record is the evidence that you evaluated the trigger.

If the activity call fails, the turn must still leave a record without breaking
the one-comment-per-turn rule. Post one short comment containing the outcome
and error **only if you have not already commented this turn**. An existing
delegation or coordination comment is already the record; never post a second
comment just to report the activity error. Silence is required only when the
no_action activity call succeeds.

## One trigger per work item

Use either a roster mention on the current issue **or** a `todo` child assigned
to the worker. Creating that assigned child already triggers the worker;
mentioning the same worker on the parent for the same work starts a duplicate
parallel run. Never do both. Prefer the roster-mention path for ordinary squad
coordination; creating sub-issues is not a requirement of this skill.

For a correction to an already-running worker, open the installed platform
skill's mentions reference and follow its steering contract. A normal comment
does not reach a running turn immediately. New work remains a normal delegation.

## Parent status and wrap-up

Current runtime permissions decide whether you may change parent status. Guest
leaders never change it; a moved workflow binding forbids all status changes.
Owners open `references/execution.md` in the installed issue-execution skill
directory before status decisions. Without a project workflow, set the parent to `in_progress`
on first assignment and keep it there while members work. On a later member-update
or stage-barrier turn that verifies the overall goal, move it to `in_review`
even without another comment asking you to; leave `done` to human review or
existing integrations. With a project workflow, leave the current step unchanged
on dispatch and run its Step done command once the current step’s goal is met,
even without another request. User definitions and permission restrictions still
apply. Roster or project membership alone grants no status authority.

When wrapping up, verify acceptance against member results, summarize the outcome
and any remaining gaps, and record activity. A successful dispatch is not parent
completion. Follow user-defined review and handoff requirements before closing
the coordination loop.
