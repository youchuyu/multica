package service

// ScopeExecutionSkills applies the same resolved-turn scope to both wire formats.
func ScopeExecutionSkills(skills []AgentSkillData, refs []AgentSkillRefData, issueTurn, leaderTurn bool) ([]AgentSkillData, []AgentSkillRefData, int, int) {
	removedInline, removedRefs := 0, 0
	inline := skills[:0]
	for _, skill := range skills {
		if (skill.Source == "builtin" || (skill.Source == "" && skill.ID == "")) && !ExecutionSkillApplies(skill.Name, issueTurn, leaderTurn) {
			removedInline++
			continue
		}
		inline = append(inline, skill)
	}
	skills = inline
	references := refs[:0]
	for _, ref := range refs {
		if ref.Source == "builtin" && ((ref.ID == BuiltinSkillID(IssueExecutionSkillName) && !issueTurn) || (ref.ID == BuiltinSkillID(SquadCoordinationSkillName) && !(issueTurn && leaderTurn))) {
			removedRefs++
			continue
		}
		references = append(references, ref)
	}
	refs = references
	return skills, refs, removedInline, removedRefs
}

// ExecutionSkillApplies scopes platform execution protocols to the resolved turn.
// Non-builtin skills are never filtered by display name.
func ExecutionSkillApplies(name string, issue, leader bool) bool {
	switch name {
	case IssueExecutionSkillName:
		return issue
	case SquadCoordinationSkillName:
		return issue && leader
	default:
		return true
	}
}
