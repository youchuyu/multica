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
	for _, template := range []struct{ header, hard, separator string }{
		{"combined-v1-header", "combined-v1-hard-rules", "\n"},
		{"canonical-v1-header", "canonical-v1-hard-rules", "\n"},
		{"canonical-v1-skill-header", "", "\n\n"},
	} {
		header := read(template.header)
		for _, status := range statuses {
			for _, steer := range []string{"", "\n" + read("steer")} {
				old := header + template.separator + status
				if template.hard != "" {
					old += "\n\n" + read(template.hard)
				}
				old += steer
				// Only a complete protocol next to the claimed leader's roster is eligible.
				needle := old + "\n\n## Squad Roster\n\nLeader (you):\n- "
				at := strings.Index(text, needle)
				if at < 0 {
					needle = old + "\n\n## Squad Roster\n\nRoster facts: members, roles, skills, and copyable mention syntax; not additional authority.\n\nLeader (you):\n- "
					at = strings.Index(text, needle)
				}
				if at < 0 || strings.Count(text, header) != 1 || (at != 0 && !strings.HasSuffix(text[:at], "\n\n")) {
					continue
				}
				row := text[at+len(needle):]
				if end := strings.IndexByte(row, '\n'); end >= 0 {
					row = row[:end]
				}
				if !strings.HasSuffix(row, "(mention://agent/"+leaderID+")"+"`") {
					continue
				}
				matches = append(matches, match{old, LocalLeaderHeader + "\n\n" + status})
			}
		}
	}
	if len(matches) != 1 {
		return text, false
	}
	return strings.Replace(text, matches[0].old, matches[0].replacement, 1), true
}
