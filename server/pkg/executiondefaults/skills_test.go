package executiondefaults

import (
	"reflect"
	"strings"
	"testing"

	"github.com/multica-ai/multica/server/pkg/skillbundle"
)

func TestSkillsEmbeddedAndIndependent(t *testing.T) {
	first, second := Skills(), Skills()
	if len(first) != 3 || !reflect.DeepEqual(first, second) {
		t.Fatal("defaults must contain exactly three deterministic skills")
	}
	for _, skill := range first {
		if !Owns(skill.Source, skill.ID) || skill.ID != "builtin:"+skill.Name || skill.Content == "" {
			t.Fatalf("invalid builtin: %+v", skill)
		}
		if skill.Name == PlatformSkillName {
			found := false
			for _, file := range skill.Files {
				if file.Path == "references/issues.md" {
					found = true
				}
			}
			if !found {
				t.Fatal("platform references missing")
			}
		}
		for _, file := range skill.Files {
			if file.Content == "" || strings.HasPrefix(file.Path, "/") || strings.Contains(file.Path, "..") {
				t.Fatalf("invalid supporting file %s", file.Path)
			}
		}
		if skillbundle.BuildManifest(skill).Hash == "" {
			t.Fatal("manifest missing")
		}
	}
	first[0].Content = "modified"
	first[0].Files[0].Content = "modified"
	if !reflect.DeepEqual(second, Skills()) {
		t.Fatal("caller mutated shared defaults")
	}
}

func TestMigratedProceduresHaveSkillOwners(t *testing.T) {
	byName := make(map[string]string)
	for _, skill := range Skills() {
		byName[skill.Name] = skill.Content
		for _, file := range skill.Files {
			byName[skill.Name+"/"+file.Path] = file.Content
		}
	}
	for path, rules := range map[string][]string{
		SquadCoordinationSkillName:                         {"no_action", "one-comment-per-turn", "current step’s goal", "even without another request", "references/execution.md", "steering contract"},
		IssueExecutionSkillName + "/references/context.md": {"source_context", "only", "server-computed issue-wide delta"},
		PlatformSkillName + "/references/runtimes.md":      {"reported daemon PID", "multica daemon status", "multica env stop"},
		PlatformSkillName + "/references/issues.md":        {"--content-file", "posting exit status", "private working copy"},
		PlatformSkillName + "/references/mentions.md":      {"Silence ends conversations", "--steer", "mention://"},
	} {
		for _, rule := range rules {
			if !strings.Contains(byName[path], rule) {
				t.Errorf("%s missing migrated rule %q", path, rule)
			}
		}
	}
	for path, body := range byName {
		for _, old := range []string{"runtime brief's Mentions section", "runtime's steering", "runtime's Definition Boundaries", "existing Squad Operating Protocol"} {
			if strings.Contains(body, old) {
				t.Errorf("%s depends on removed runtime procedure %q", path, old)
			}
		}
	}
}

func TestPlatformReferencesExcludeGeneralHandoffTutorials(t *testing.T) {
	for _, skill := range Skills() {
		for _, file := range skill.Files {
			if skill.Name != PlatformSkillName || file.Path != "references/runtimes.md" {
				continue
			}
			for _, generic := range []string{"foreground blocking watch", "durable logs", "unsupervised survival", "repository merge gate"} {
				if strings.Contains(file.Content, generic) {
					t.Errorf("general tutorial retained in platform reference: %q", generic)
				}
			}
		}
	}
}

func TestOwnsRequiresExplicitBuiltinIdentity(t *testing.T) {
	id := BuiltinSkillID(PlatformSkillName)
	for _, pair := range [][2]string{{"workspace", id}, {"plugin", id}, {"", id}, {"builtin", PlatformSkillName}, {"builtin", "builtin:multica-onboarding"}} {
		if Owns(pair[0], pair[1]) {
			t.Fatalf("unexpected owned identity %v", pair)
		}
	}
}
