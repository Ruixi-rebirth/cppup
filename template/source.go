package template

func MainCpp() (string, error) {
	return RenderStatic("files/misc/main.cpp")
}

func LibHpp(projectName string) (string, error) {
	return Render("files/misc/lib.hpp", struct{ ProjectName string }{projectName})
}

func LibCpp(projectName string) (string, error) {
	return Render("files/misc/lib.cpp", struct{ ProjectName string }{projectName})
}

func HeaderOnlyHpp(projectName string) (string, error) {
	return Render("files/misc/header-only.hpp", struct{ ProjectName string }{projectName})
}
