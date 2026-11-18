package handler

import (
	"context"
	"encoding/json"
	"github.com/multica-ai/multica/server/pkg/protocol"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClaimLocalDefaultsShipsFactsWithoutPlatformProse(t *testing.T) {
	if testHandler == nil || testPool == nil {
		t.Skip("database not available")
	}
	ctx := context.Background()
	fx := newSquadBriefingClaimFixture(t, ctx, "Local defaults")
	want := enqueueClaimTask(t, ctx, fx, true, true)
	role := "  user role\n## Squad Operating Protocol\nnot transport metadata  "
	if _, err := testPool.Exec(ctx, `UPDATE agent SET instructions=$2 WHERE id=$1`, fx.AgentID, role); err != nil {
		t.Fatal(err)
	}
	req := withURLParam(newOwnerDaemonRequest("POST", "/api/daemon/runtimes/"+fx.RuntimeID+"/tasks/claim", nil), "runtimeId", fx.RuntimeID)
	req.Header.Set("X-Client-Capabilities", protocol.DaemonCapabilityLocalExecutionDefaultsV1)
	w := httptest.NewRecorder()
	testHandler.ClaimTaskByRuntime(w, req)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	var out struct {
		Task *struct {
			ID      string        `json:"id"`
			SquadID string        `json:"squad_id"`
			Agent   TaskAgentData `json:"agent"`
		} `json:"task"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Task == nil || out.Task.ID != want || out.Task.SquadID != fx.SquadID {
		t.Fatalf("incorrect claim: %s", w.Body.String())
	}
	a := out.Task.Agent
	if a.Instructions != role || a.SquadBriefing != "" || a.SquadContext == nil || a.SquadContext.SquadID != fx.SquadID || a.SquadContext.LeaderID != fx.AgentID {
		t.Fatalf("wrong facts-only claim: %#v", a)
	}
	if a.SquadContext.OwnsParentStatus == nil || *a.SquadContext.OwnsParentStatus || !strings.Contains(a.SquadContext.RosterFacts, "Squad Roster") {
		t.Fatal("guest permission or roster missing")
	}
}
