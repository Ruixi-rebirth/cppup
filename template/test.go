package template

import "strings"

// archiveURL computes a GitHub tarball URL from a git repo URL and tag.
func archiveURL(gitRepo, tag string) string {
	base := strings.TrimSuffix(gitRepo, ".git")
	return base + "/archive/refs/tags/" + tag + ".tar.gz"
}

// TestCMakeLists generates tests/CMakeLists.txt
func TestCMakeLists(projectName, projType, framework, version string) (string, error) {
	fw, _ := TFrameworkByName(framework)
	libName := ""
	if projType != "exe" {
		libName = projectName
	}
	return Render("files/tests/cmake/"+framework+".cmake", struct {
		ArchiveURL string
		LibName    string
	}{archiveURL(fw.GitRepo, version), libName})
}

// TestMesonBuild generates tests/meson.build
func TestMesonBuild(projectName, projType, framework string) (string, error) {
	libName := ""
	if projType != "exe" {
		libName = projectName
	}
	return Render("files/tests/meson/"+framework+".meson", struct {
		LibName string
	}{libName})
}

// TestSourceCpp generates tests/test_main.cpp
func TestSourceCpp(framework string) (string, error) {
	return RenderStatic("files/tests/cpp/" + framework + ".cpp")
}

// MesonWrap generates a subprojects/*.wrap file
func MesonWrap(framework, version string) (string, error) {
	fw, ok := TFrameworkByName(framework)
	if !ok {
		return "", nil
	}
	return Render("files/wrap/"+framework+".wrap", struct {
		GitRepo string
		Version string
	}{fw.GitRepo, version})
}

func MesonWrapFilename(framework string) string {
	switch framework {
	case "googletest":
		return "gtest.wrap"
	case "catch2":
		return "catch2.wrap"
	case "doctest":
		return "doctest.wrap"
	}
	return ""
}
