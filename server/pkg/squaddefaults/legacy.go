package squaddefaults

import (
	"embed"
	"strings"
)

// These immutable signatures are the complete generated template observed in
// legacy combined claims (MUL-277), not headings or user-declaration delimiters.
// Keep them frozen when changing the canonical defaults in defaults.go.
//
//go:embed legacy/*.txt
var legacyTemplates embed.FS

// ReplaceLegacyProtocol replaces only a complete known generated protocol.
// Status authority is matched against a finite set of explicit server statements,
// then preserved verbatim; roster/user content and every byte outside the match
// are untouched. Unknown variants deliberately remain unchanged, without errors.
func ReplaceLegacyProtocol(text, leaderID string) (string, bool) {
	if leaderID == "" {
		return text, false
	}
	header, _ := legacyTemplates.ReadFile("legacy/combined-v1-header.txt")
	hard, _ := legacyTemplates.ReadFile("legacy/combined-v1-hard-rules.txt")
	read := func(name string) string {
		b, err := legacyTemplates.ReadFile("legacy/" + name + ".txt")
		if err != nil {
			panic(err)
		}
		return string(b)
	}
	statuses := []string{read("owned"), read("guest"), read("workflow"), read("workflow") + read("workflow-override"), read("moved")}
	type match struct{ old, replacement string }
	var matches []match
	for _, status := range statuses {
		for _, steer := range []string{"", "\n" + read("steer")} {
			old := string(header) + "\n" + status + "\n\n" + string(hard) + steer
			// The protocol must be a complete block immediately followed by the
			// transport's roster with this claimed leader, not an isolated quotation.
			needle := old + "\n\n## Squad Roster\n\nLeader (you):\n- "
			at := strings.Index(text, needle)
			if at < 0 || strings.Count(text, string(header)) != 1 || (at != 0 && !strings.HasSuffix(text[:at], "\n\n")) {
				continue
			}
			row := text[at+len(needle):]
			if end := strings.IndexByte(row, '\n'); end >= 0 {
				row = row[:end]
			}
			if !strings.HasSuffix(row, "(mention://agent/"+leaderID+")"+"`") {
				continue
			}
			matches = append(matches, match{old, SquadOperatingProtocolHeader + "\n" + status + "\n\n" + SquadOperatingProtocolHardRules + steer})
		}
	}
	if len(matches) != 1 {
		return text, false
	}
	return strings.Replace(text, matches[0].old, matches[0].replacement, 1), true
}
