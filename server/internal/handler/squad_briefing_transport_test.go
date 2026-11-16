package handler

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAttachSquadBriefingPreservesPayloads(t *testing.T) {
	const briefing = "## Squad Operating Protocol\nOriginal protocol\n\n## Squad Roster\nOriginal roster\n\n## Squad Instructions (team)\nOriginal instructions  \n"
	for _, instructions := range []string{"", " \n", "  Agent instructions\n## Squad Operating Protocol\nAuthored heading\n  "} {
		for _, separate := range []bool{false, true} {
			agent := &TaskAgentData{Instructions: instructions}
			attachSquadBriefing(agent, briefing, separate)
			if separate {
				if agent.Instructions != instructions || agent.SquadBriefing != briefing {
					t.Fatal("separate payloads changed")
				}
			} else {
				want := briefing
				if strings.TrimSpace(instructions) != "" {
					want = instructions + "\n\n" + briefing
				}
				if agent.Instructions != want || agent.SquadBriefing != "" {
					t.Fatal("legacy wire contract changed")
				}
			}
			raw, err := json.Marshal(agent)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "squad_briefing") != separate {
				t.Fatal("briefing field must be omitted for legacy clients")
			}
		}
	}
}
