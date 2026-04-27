package cmd

import (
	"encoding/json"
	"os"
)

const metaFile = ".cppup"

type ProjectMeta struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	BuildSystem   string `json:"build_system"`
	Version       string `json:"version"`
	Std           string `json:"std"`
	TestFramework string `json:"test_framework,omitempty"`
}

func readMeta() (ProjectMeta, error) {
	data, err := os.ReadFile(metaFile)
	if err != nil {
		return ProjectMeta{}, err
	}
	var m ProjectMeta
	return m, json.Unmarshal(data, &m)
}

func isExe() bool {
	m, err := readMeta()
	if err != nil {
		return true
	}
	return m.Type == "exe"
}

func hasTests() bool {
	m, err := readMeta()
	if err != nil {
		return false
	}
	return m.TestFramework != ""
}
