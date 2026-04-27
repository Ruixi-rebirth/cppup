package template

func ClangFormat() (string, error) {
	return RenderStatic("files/misc/clang-format")
}

func ClangTidy() (string, error) {
	return RenderStatic("files/misc/clang-tidy")
}
