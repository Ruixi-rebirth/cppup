package cmd

import (
	"testing"

	"github.com/Ruixi-rebirth/cppup/template"
)

func TestValidateSettings_valid(t *testing.T) {
	cfg := ProjectConfig{
		Type:    "exe",
		Build:   "cmake",
		Std:     "17",
		Version: "0.1.0",
	}
	if err := validateSettings(cfg); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSettings_invalidType(t *testing.T) {
	cfg := ProjectConfig{Type: "bin", Build: "cmake", Std: "17"}
	if err := validateSettings(cfg); err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestValidateSettings_validTypes(t *testing.T) {
	for _, typ := range []string{"exe", "lib-static", "lib-shared", "lib-header"} {
		cfg := ProjectConfig{Type: typ, Build: "cmake", Std: "17"}
		if err := validateSettings(cfg); err != nil {
			t.Errorf("type %q should be valid: %v", typ, err)
		}
	}
}

func TestValidateSettings_invalidBuild(t *testing.T) {
	cfg := ProjectConfig{Type: "exe", Build: "make", Std: "17"}
	if err := validateSettings(cfg); err == nil {
		t.Error("expected error for invalid build system")
	}
}

func TestValidateSettings_validBuilds(t *testing.T) {
	for _, bs := range template.BuildSystems {
		cfg := ProjectConfig{Type: "exe", Build: bs.Name, Std: "17"}
		if err := validateSettings(cfg); err != nil {
			t.Errorf("build %q should be valid: %v", bs.Name, err)
		}
	}
}

func TestValidateSettings_invalidStd(t *testing.T) {
	cfg := ProjectConfig{Type: "exe", Build: "cmake", Std: "03"}
	if err := validateSettings(cfg); err == nil {
		t.Error("expected error for invalid std")
	}
}

func TestValidateSettings_validStds(t *testing.T) {
	for _, std := range []string{"11", "14", "17", "20", "23"} {
		cfg := ProjectConfig{Type: "exe", Build: "cmake", Std: std}
		if err := validateSettings(cfg); err != nil {
			t.Errorf("std %q should be valid: %v", std, err)
		}
	}
}

func TestValidateSettings_invalidFramework(t *testing.T) {
	cfg := ProjectConfig{Type: "exe", Build: "cmake", Std: "17", Tests: "boost"}
	if err := validateSettings(cfg); err == nil {
		t.Error("expected error for invalid test framework")
	}
}

func TestValidateSettings_validFrameworks(t *testing.T) {
	for _, fw := range template.TFrameworks {
		cfg := ProjectConfig{Type: "exe", Build: "cmake", Std: "17", Tests: fw.Name}
		if err := validateSettings(cfg); err != nil {
			t.Errorf("framework %q should be valid: %v", fw.Name, err)
		}
	}
}

func TestValidateSettings_noTests(t *testing.T) {
	cfg := ProjectConfig{Type: "exe", Build: "cmake", Std: "17", Tests: ""}
	if err := validateSettings(cfg); err != nil {
		t.Errorf("empty Tests should be valid: %v", err)
	}
}
