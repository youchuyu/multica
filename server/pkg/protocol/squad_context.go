package protocol

const SquadContextSchemaV1 = 1

// SquadWorkflow is server-resolved permission context, not generated protocol prose.
type SquadWorkflow string

const (
	SquadWorkflowNone      SquadWorkflow = "none"
	SquadWorkflowProject   SquadWorkflow = "project"
	SquadWorkflowStepMoved SquadWorkflow = "step_moved"
)

// SquadContext carries server-owned facts only. Default coordination rules belong
// to the installed daemon. Pointer permission distinguishes false from omission.
type SquadContext struct {
	SchemaVersion    int    `json:"schema_version"`
	SquadID          string `json:"squad_id"`
	SquadName        string `json:"squad_name"`
	LeaderID         string `json:"leader_id"`
	UserInstructions string `json:"user_instructions"`
	// RosterFacts is rendered member/role/skill/mention data, never default protocol.
	RosterFacts      string        `json:"roster_facts"`
	OwnsParentStatus *bool         `json:"owns_parent_status"`
	Workflow         SquadWorkflow `json:"workflow"`
}
