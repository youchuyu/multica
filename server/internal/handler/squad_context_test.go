package handler

import (
	"encoding/json"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
	"strings"
	"testing"
)

func TestNewSquadContextContainsFactsNotDefaults(t *testing.T) {
	squad := db.Squad{ID: parseUUID("11111111-1111-1111-1111-111111111111"), LeaderID: parseUUID("22222222-2222-2222-2222-222222222222"), Name: "Team", Instructions: "  exact instructions\n## Custom heading\n"}
	for _, tc := range []struct {
		in   leaderWorkflow
		want protocol.SquadWorkflow
	}{{leaderNoWorkflow, protocol.SquadWorkflowNone}, {leaderWorkflowCurrentDaemon, protocol.SquadWorkflowProject}, {leaderWorkflowLegacyDaemon, protocol.SquadWorkflowProject}, {leaderWorkflowStepMoved, protocol.SquadWorkflowStepMoved}} {
		facts := newSquadContext(squad, "roster facts", false, tc.in)
		if facts.SchemaVersion != 1 || facts.SquadID != "11111111-1111-1111-1111-111111111111" || facts.LeaderID != "22222222-2222-2222-2222-222222222222" || facts.UserInstructions != squad.Instructions || facts.RosterFacts != "roster facts" || facts.OwnsParentStatus == nil || *facts.OwnsParentStatus || facts.Workflow != tc.want {
			t.Fatalf("incorrect facts: %#v", facts)
		}
		wire, err := json.Marshal(facts)
		if err != nil {
			t.Fatal(err)
		}
		for _, bad := range []string{"Generated coordination defaults", "Coordinate, do not implement", "Record every trigger"} {
			if strings.Contains(string(wire), bad) {
				t.Fatalf("default prose leaked into facts: %s", wire)
			}
		}
		if !strings.Contains(string(wire), "\"owns_parent_status\":false") {
			t.Fatal("false permission must remain explicit")
		}
	}
}
