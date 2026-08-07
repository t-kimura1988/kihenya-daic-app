package util

import (
	"strings"
	"testing"
)

func TestGenerateIdealCode(t *testing.T) {
	code, err := GenerateIdealCode()
	if err != nil {
		t.Fatalf("GenerateIdealCode: %v", err)
	}
	if !strings.HasPrefix(code, "il_") {
		t.Errorf("code %q does not have il_ prefix", code)
	}
	if len(code) != len("il_")+17 {
		t.Errorf("code length = %d, want %d", len(code), len("il_")+17)
	}
}

func TestGenerateGoalCode(t *testing.T) {
	code, err := GenerateGoalCode()
	if err != nil {
		t.Fatalf("GenerateGoalCode: %v", err)
	}
	if !strings.HasPrefix(code, "gl_") {
		t.Errorf("code %q does not have gl_ prefix", code)
	}
	if len(code) != len("gl_")+17 {
		t.Errorf("code length = %d, want %d", len(code), len("gl_")+17)
	}
}

func TestGenerateTaskCode(t *testing.T) {
	code, err := GenerateTaskCode()
	if err != nil {
		t.Fatalf("GenerateTaskCode: %v", err)
	}
	if !strings.HasPrefix(code, "tk_") {
		t.Errorf("code %q does not have tk_ prefix", code)
	}
	if len(code) != len("tk_")+17 {
		t.Errorf("code length = %d, want %d", len(code), len("tk_")+17)
	}
}
