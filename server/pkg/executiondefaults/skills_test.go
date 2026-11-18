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

func TestOwnsRequiresExplicitBuiltinIdentity(t *testing.T) {
	id := BuiltinSkillID(PlatformSkillName)
	for _, pair := range [][2]string{{"workspace", id}, {"plugin", id}, {"", id}, {"builtin", PlatformSkillName}, {"builtin", "builtin:multica-onboarding"}} {
		if Owns(pair[0], pair[1]) {
			t.Fatalf("unexpected owned identity %v", pair)
		}
	}
}
