package template

import "strings"

func FlakeNix(projectName, buildSystem, version, testFramework string) (string, error) {
	var buildInputs []string
	switch buildSystem {
	case "meson":
		buildInputs = []string{"meson", "ninja", "pkg-config"}
	default:
		buildInputs = []string{"cmake", "ninja"}
	}
	switch testFramework {
	case "googletest":
		buildInputs = append(buildInputs, "gtest")
	case "catch2":
		buildInputs = append(buildInputs, "catch2_3")
	case "doctest":
		buildInputs = append(buildInputs, "doctest")
	}
	return Render("files/misc/flake.nix", struct {
		ProjectName       string
		Version           string
		NativeBuildInputs string
	}{projectName, version, "[ " + strings.Join(buildInputs, " ") + " ]"})
}

func TreefmtNix() (string, error) {
	return RenderStatic("files/misc/treefmt.nix")
}

func Envrc() (string, error) {
	return RenderStatic("files/misc/envrc")
}
