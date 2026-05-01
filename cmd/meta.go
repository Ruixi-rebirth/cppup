package cmd

import (
	"encoding/json"
	"os"
)

const metaFile = ".cppup"

type Dep struct {
	Name string `json:"name"`
	Git  string `json:"git,omitempty"`
	Tag  string `json:"tag,omitempty"`
	URL  string `json:"url,omitempty"`
	Hash string `json:"hash,omitempty"`
}

func (d Dep) Info() string {
	switch {
	case d.Git != "":
		return "[Git: " + d.Tag + "]"
	case d.URL != "":
		return "[URL]"
	default:
		return "[WrapDB]"
	}
}

type ProjectMeta struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	BuildSystem   string `json:"build_system"`
	Version       string `json:"version"`
	Std           string `json:"std"`
	TestFramework string `json:"test_framework,omitempty"`
	Deps          []Dep  `json:"deps,omitempty"`
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
