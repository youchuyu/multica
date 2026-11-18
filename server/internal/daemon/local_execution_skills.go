package daemon

import (
	"github.com/multica-ai/multica/server/pkg/executiondefaults"
	"github.com/multica-ai/multica/server/pkg/skillbundle"
)

// normalizeEquivalentLegacyExecutionSkills recognizes only byte-equivalent anonymous
// payloads, not provenance. User skills with a UUID or any explicit source are
// untouched. Unknown anonymous payloads are preserved as supplied.
func normalizeEquivalentLegacyExecutionSkills(task *Task) {
	if task == nil || task.Agent == nil {
		return
	}
	for i := range task.Agent.Skills {
		skill := &task.Agent.Skills[i]
		if skill.ID != "" || skill.Source != "" {
			continue
		}
		for _, local := range executiondefaults.Skills() {
			if skill.Name != local.Name || skill.Description != local.Description || skill.Content != local.Content || len(skill.Files) != len(local.Files) {
				continue
			}
			files := make(map[string]string, len(skill.Files))
			duplicate := false
			for _, file := range skill.Files {
				if _, ok := files[file.Path]; ok {
					duplicate = true
				}
				files[file.Path] = file.Content
			}
			match := !duplicate
			for _, file := range local.Files {
				content, ok := files[file.Path]
				if !ok || content != file.Content {
					match = false
					break
				}
			}
			if match {
				skill.ID, skill.Source = local.ID, local.Source
				break
			}
		}
	}
}

// filterLocalExecutionSkillRefs runs before cache/network resolution. Locally owned
// builtins never need a server hash or download, including out-of-scope entries.
// Unknown builtins and all user/plugin declarations still resolve remotely.
func filterLocalExecutionSkillRefs(task *Task) {
	if task == nil || task.Agent == nil {
		return
	}
	refs := make([]SkillRefData, 0, len(task.Agent.SkillRefs))
	for _, ref := range task.Agent.SkillRefs {
		if !executiondefaults.Owns(ref.Source, ref.ID) {
			refs = append(refs, ref)
		}
	}
	task.Agent.SkillRefs = refs
}

// mergeLocalExecutionSkills runs after remote resolution so inline-only tasks and
// tasks with bundle references receive exactly the same local execution defaults.
// Missing/ambiguous builtin identity is not permission to replace user content.
func mergeLocalExecutionSkills(task *Task) {
	if task == nil || task.Agent == nil {
		return
	}
	skills := make([]SkillData, 0, len(task.Agent.Skills)+3)
	for _, skill := range task.Agent.Skills {
		if !executiondefaults.Owns(skill.Source, skill.ID) {
			skills = append(skills, skill)
		}
	}
	for _, local := range executiondefaults.Skills() {
		if local.Name == executiondefaults.IssueExecutionSkillName && task.IssueID == "" {
			continue
		}
		if local.Name == executiondefaults.SquadCoordinationSkillName && (task.IssueID == "" || !taskIsSquadLeader(*task)) {
			continue
		}
		manifest := skillbundle.BuildManifest(local)
		skill := SkillData{ID: local.ID, Source: local.Source, Name: local.Name, Description: local.Description, Content: local.Content, Hash: manifest.Hash, SizeBytes: manifest.SizeBytes}
		refs := make(map[string]skillbundle.FileRef, len(manifest.Files))
		for _, file := range manifest.Files {
			refs[file.Path] = file
		}
		for _, file := range local.Files {
			ref := refs[file.Path]
			skill.Files = append(skill.Files, SkillFileData{Path: file.Path, Content: file.Content, SHA256: ref.SHA256, SizeBytes: ref.SizeBytes})
		}
		skills = append(skills, skill)
	}
	task.Agent.Skills = skills
}
