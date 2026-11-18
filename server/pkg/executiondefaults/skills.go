// Package executiondefaults owns the execution contracts shipped with a local build.
// Server-owned declarations, task facts and permissions are not defaults.
package executiondefaults

import (
	"embed"
	"io/fs"
	"path"
	"strings"

	"github.com/multica-ai/multica/server/pkg/skillbundle"
)

const (
	PlatformSkillName          = "multica-platform"
	IssueExecutionSkillName    = "multica-issue-execution"
	SquadCoordinationSkillName = "multica-squad-coordination"
)

//go:embed skills
var skillsFS embed.FS

// BuiltinSkillID preserves the existing wire identity, independently of allocated names.
func BuiltinSkillID(name string) string { return "builtin:" + name }

// Owns matches only explicit builtin identities, never a user skill's display name.
func Owns(source, id string) bool {
	if source != skillbundle.SourceBuiltin {
		return false
	}
	switch id {
	case BuiltinSkillID(PlatformSkillName), BuiltinSkillID(IssueExecutionSkillName), BuiltinSkillID(SquadCoordinationSkillName):
		return true
	}
	return false
}

// Skills returns fresh skill values and supporting files in deterministic name order.
// System-agent-only procedures such as onboarding intentionally remain server-owned.
func Skills() []skillbundle.Skill {
	names := []string{IssueExecutionSkillName, PlatformSkillName, SquadCoordinationSkillName}
	skills := make([]skillbundle.Skill, 0, len(names))
	for _, name := range names {
		dir := path.Join("skills", name)
		content, err := fs.ReadFile(skillsFS, path.Join(dir, "SKILL.md"))
		if err != nil {
			panic("executiondefaults: " + err.Error())
		}
		skill := skillbundle.Skill{ID: BuiltinSkillID(name), Source: skillbundle.SourceBuiltin, Name: name, Content: string(content)}
		err = fs.WalkDir(skillsFS, dir, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || p == path.Join(dir, "SKILL.md") {
				return nil
			}
			data, err := fs.ReadFile(skillsFS, p)
			if err != nil {
				return err
			}
			skill.Files = append(skill.Files, skillbundle.File{Path: strings.TrimPrefix(p, dir+"/"), Content: string(data)})
			return nil
		})
		if err != nil {
			panic("executiondefaults: " + err.Error())
		}
		skills = append(skills, skill)
	}
	return skills
}
