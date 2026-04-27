package template

func MesonBuildExe(projectName, version, std, minVersion string, withTests bool) (string, error) {
	return Render("files/meson/exe.meson", struct {
		ProjectName string
		Version     string
		CxxStd      string
		MinVersion  string
		WithTests   bool
	}{projectName, version, std, minVersion, withTests})
}

func MesonBuildLib(projectName, version, std, libType, minVersion string, withTests bool) (string, error) {
	libFn := "static_library"
	if libType == "shared" {
		libFn = "shared_library"
	}
	return Render("files/meson/lib.meson", struct {
		ProjectName string
		Version     string
		CxxStd      string
		LibFn       string
		MinVersion  string
		WithTests   bool
	}{projectName, version, std, libFn, minVersion, withTests})
}

func MesonBuildHeaderOnly(projectName, version, std, minVersion string, withTests bool) (string, error) {
	return Render("files/meson/header-only.meson", struct {
		ProjectName string
		Version     string
		CxxStd      string
		MinVersion  string
		WithTests   bool
	}{projectName, version, std, minVersion, withTests})
}
