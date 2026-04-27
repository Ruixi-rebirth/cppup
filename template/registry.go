package template

// TFramework describes a supported test framework.
type TFramework struct {
	Name       string
	DefaultTag string
	GitRepo    string
}

// TFrameworks is the ordered list of supported test frameworks.
var TFrameworks = []TFramework{
	{"googletest", "v1.17.0", "https://github.com/google/googletest.git"},
	{"catch2", "v3.14.0", "https://github.com/catchorg/Catch2.git"},
	{"doctest", "v2.5.2", "https://github.com/doctest/doctest.git"},
}

// TFrameworkByName returns the TFramework with the given name.
func TFrameworkByName(name string) (TFramework, bool) {
	for _, f := range TFrameworks {
		if f.Name == name {
			return f, true
		}
	}
	return TFramework{}, false
}

// BuildSystem describes a supported build system.
type BuildSystem struct {
	Name       string
	BuildFile  string
	MinVersion string
}

// BuildSystems is the ordered list of supported build systems.
var BuildSystems = []BuildSystem{
	{"cmake", "CMakeLists.txt", "3.20"},
	{"meson", "meson.build", "1.0"},
}

// BuildSystemByName returns the BuildSystem with the given name.
func BuildSystemByName(name string) (BuildSystem, bool) {
	for _, b := range BuildSystems {
		if b.Name == name {
			return b, true
		}
	}
	return BuildSystem{}, false
}
