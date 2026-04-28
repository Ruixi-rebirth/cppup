package template

func CMakeListsExe(projectName, version, std, minVersion string, withTests bool) (string, error) {
	return Render("files/cmake/exe.cmake", struct {
		ProjectName string
		Version     string
		CxxStd      string
		MinVersion  string
		WithTests   bool
	}{projectName, version, std, minVersion, withTests})
}

func CMakeListsLib(projectName, version, std, libType, minVersion string, withTests bool) (string, error) {
	return Render("files/cmake/lib.cmake", struct {
		ProjectName string
		Version     string
		CxxStd      string
		LibType     string
		MinVersion  string
		WithTests   bool
	}{projectName, version, std, libType, minVersion, withTests})
}

func CMakeListsHeaderOnly(projectName, version, minVersion string, withTests bool) (string, error) {
	return Render("files/cmake/header-only.cmake", struct {
		ProjectName string
		Version     string
		MinVersion  string
		WithTests   bool
	}{projectName, version, minVersion, withTests})
}

func CMakeConfigIn() (string, error) {
	return RenderStatic("files/cmake/config.cmake.in")
}

func CMakePCIn(headerOnly bool) (string, error) {
	if headerOnly {
		return RenderStatic("files/cmake/header-only.pc.in")
	}
	return RenderStatic("files/cmake/lib.pc.in")
}

func Gitignore(nix, meson bool) (string, error) {
	return Render("files/misc/gitignore", struct {
		Nix   bool
		Meson bool
	}{nix, meson})
}
