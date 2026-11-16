---
name: multica-issue-execution
description: "Required for every Multica Issue turn: assignments, issue comments, issue wakeups and squad issue runs. Read issue context and valid read waivers, execute within role boundaries, deliver results and maintain issue/workflow status. Not for chat, quick-create or run-only autopilot tasks."
user-invocable: false
allowed-tools: Bash(multica *), Bash(git *), Bash(gh *)
---

# Execute an Issue Turn

This skill governs **Issue tasks only**, not chat (even chat about an issue),
quick-create or run-only autopilot. Before any issue action, read **both**:

1. [Context and read waivers](references/context.md)
2. [Execution, delivery and status](references/execution.md)

Agent Identity and the runtime's Instruction Precedence always win. This skill
never widens your role or grants permission to investigate, implement, create,
update, delegate, comment or change status when that action is forbidden.
Keep the current role instructions, including the Squad Operating Protocol,
unchanged. A delegation-only role stops once its allowed delegation is delivered.

The per-turn user message supplies the issue ID, trigger, reply threads,
context-read commands and any `[Project workflow]`. Those are run facts, not
skill defaults: never reuse a previous turn's values.

The runtime's Output section governs the delivery surface. Detailed CLI delivery
contracts (file-first comments, receipts, attachments, PR links, charts) remain
in the installed **multica-platform** skill and its issues reference; open that
reference before delivery. Use the allocated installed name in the runtime's
skill list if a built-in name was renamed. Invoke all CLI commands through the
runtime's quoted `MULTICA_CLI_PATH`, never a PATH lookup.
